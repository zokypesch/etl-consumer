package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/big"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/kennygrant/sanitize"
	"github.com/zokypesch/etl/client"
	"github.com/zokypesch/etl/config"
	"github.com/zokypesch/etl/data"
	"github.com/zokypesch/etl/scheme"
	"github.com/zokypesch/etl/utils"
	"github.com/zokypesch/proto-lib/core"
	"gopkg.in/confluentinc/confluent-kafka-go.v1/kafka"
)

func main() {

	cfg := config.Get()
	db := core.InitDB(cfg.DBAddress, cfg.DBSourceName, cfg.DBUser, cfg.DBPassword, cfg.DBPort, cfg.DBLog)

	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": cfg.Kafka,
		"group.id":          cfg.Group,
		"auto.offset.reset": cfg.AutoOffset, // earliest for beginning
	})

	if err != nil {
		panic(err)
	}

	// republish message
	p, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": cfg.Kafka})

	if err != nil {
		panic(err)
	}

	api := client.NewAPI(cfg.DebeziumAddr, cfg.DebeziumPort, cfg.Connector)

	/* rds internal mark */
	var listTopic []string
	for _, vTable := range cfg.Table {
		listTopic = append(listTopic, fmt.Sprintf("%s.%s.%s", cfg.Server, cfg.DBName, vTable))
	}

	scheme := fmt.Sprintf("%s", cfg.Server)
	if cfg.ActiveScheme {
		listTopic = []string{scheme}
	}

	c.SubscribeTopics(listTopic, nil)
	log.Printf("starting subscribe %s.%s.%s in scheme of: %s", cfg.Server, cfg.DBName, cfg.Table, scheme)

	for {
		msg, err := c.ReadMessage(-1)
		if err != nil {
			log.Printf("Consumer error: %v (%v)\n", err, msg)
			continue
		}

		topicSource := *msg.TopicPartition.Topic
		if topicSource == scheme {
			qryScheme, execute, err := processScheme(msg.Value, cfg.Table, api, cfg.ReplaceAllScheme, cfg.Reclaim)

			if err != nil {
				log.Println("schema erorr: ", err.Error())
				errLogScheme := db.Exec(fmt.Sprintf("INSERT INTO data_err (data, error, `table_name`, `db_name`) VALUES('%s', '%s', '%s', '%s')", string(msg.Value), sanitize.BaseName(err.Error()), cfg.Table, cfg.DBName)).Error

				if errLogScheme != nil {
					log.Println("failed to insert log error", errLogScheme)
				}
			}

			if !execute {
				continue
			}

			qryScheme = strings.Replace(qryScheme, fmt.Sprintf("`%s`.", cfg.DBName), "", -1)
			err = db.Exec(qryScheme).Error

			if err != nil {
				log.Println("schema erorr: ", err.Error())
				errLogScheme := db.Exec(fmt.Sprintf("INSERT INTO data_err (data, error, `table_name`, `db_name`) VALUES('%s', '%s', '%s', '%s')", string(msg.Value), sanitize.BaseName(err.Error()), cfg.Table, cfg.DBName)).Error

				if errLogScheme != nil {
					log.Println("failed to insert log error", errLogScheme)
				}

				if cfg.Republish {
					count := 1
					for _, headerMsg := range msg.Headers {
						if headerMsg.Key == "loop" {
							loop := string(headerMsg.Value)

							num, err := strconv.Atoi(loop)
							if err == nil {
								count = num + 1
							}
						}
					}

					if count < cfg.RepublishLimit {
						log.Println("republish message count: ", count)
						publish(topicSource, p, msg.Value, []byte(strconv.Itoa(count)), cfg)
					}
				}
			}
			log.Println("change scheme success: ", qryScheme)

			// resume sync process
			resume(api)
			continue
		}

		if len(string(msg.Value)) == 0 {
			continue
		}

		qry, errQry := processData(msg.Value, cfg)

		if errQry != nil {
			log.Println("failed parse json: ", errQry, string(msg.Value))
			continue
		}

		err = db.Exec(qry).Error

		if err != nil {
			if strings.Contains(err.Error(), "Duplicate entry") {
				// skip duplicate entryy
				continue
			}
			log.Println("error exec qry :", err.Error())
			errLog := db.Exec(fmt.Sprintf("INSERT INTO data_err (data, error, `table_name`, `db_name`) VALUES('%s', '%s', '%s', '%s')", string(msg.Value), sanitize.BaseName(err.Error()), cfg.Table, cfg.DBName)).Error

			if errLog != nil {
				log.Println("failed to insert log error", errLog)
			}

			if cfg.Republish {
				// check loop number
				count := 1
				for _, headerMsg := range msg.Headers {
					if headerMsg.Key == "loop" {
						loop := string(headerMsg.Value)
						num, err := strconv.Atoi(loop)
						if err == nil {
							count = num + 1
						}
					}
				}

				if count < cfg.RepublishLimit {
					log.Println("republish message count: ", count)
					publish(topicSource, p, msg.Value, []byte(strconv.Itoa(count)), cfg)
				}

			}
		}
	}

	c.Close()
}

func publish(topicSource string, p *kafka.Producer, val []byte, count []byte, cfg *config.Config) {
	// republish message
	newProducer, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": cfg.Kafka})

	log.Println("republish: ", topicSource)
	// wait for 100ms
	time.Sleep(time.Millisecond * 100)
	deliveryChan := make(chan kafka.Event)

	err = newProducer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topicSource, Partition: kafka.PartitionAny},
		Value:          val,
		Headers:        []kafka.Header{{Key: "loop", Value: count}},
	}, deliveryChan)

	if err != nil {
		log.Println("err publish: ", err.Error())
	}
	e := <-deliveryChan
	m := e.(*kafka.Message)

	if m.TopicPartition.Error != nil {
		log.Printf("Republish message failed: %v\n", m.TopicPartition.Error)
	} else {
		log.Printf("Republish message to topic %s [%d] at offset %v\n",
			*m.TopicPartition.Topic, m.TopicPartition.Partition, m.TopicPartition.Offset)
	}

	close(deliveryChan)
}

func mapToString(param map[string]interface{}, fields []data.Field,
	action string, cfg *config.Config) (string, string, string, string) {
	var key []string
	var val []string
	var comb []string

	afterField := data.SearchFieldByName(fields, action)
	re := regexp.MustCompile("((19|20)\\d\\d)-(0?[1-9]|1[012])-(0?[1-9]|[12][0-9]|3[01])")
	// re_decimal := regexp.MustCompile(`/^\d+\.?\d*$/`)

	for k, v := range param {
		if v == nil {
			continue
		}
		var p string
		key = append(key, fmt.Sprintf("`%s`", k))

		field := afterField.SearchFieldsByName(k)
		switch field.Name {
		case "org.apache.kafka.connect.data.Decimal":
			dataC, err := base64.StdEncoding.DecodeString(v.(string))
			if err != nil {
				p = fmt.Sprintf("'%d'", 0)
			} else {
				out := new(big.Int).SetBytes(dataC)
				if len(dataC) > 0 && dataC[0]&0x80 != 0 {
					// It's negative.
					// Convert 2's complement negative to abs big-endian:
					data2 := make([]byte, len(dataC)+1)
					data2[0] = 1
					temp := new(big.Int).SetBytes(data2)
					out.Sub(temp, out)

					// Apply negative sign:
					out.Neg(out)
				}

				finalScale := float64(1)
				// finalPrec := float64(0)

				scale, err := strconv.Atoi(field.Parameters.Scale)
				if err == nil {
					finalScale = float64(scale)
				}

				// prec, err := strconv.Atoi(field.Parameters.DecimalPrecision)
				// if err == nil {
				// 	finalPrec = float64(prec)
				// }

				finalResult := float64(out.Int64()) / math.Pow(10, finalScale)

				p = fmt.Sprintf("'%.2f'", finalResult)
			}
		case "io.debezium.time.Date":
			i := v.(float64)
			t, _ := time.Parse("2006-01-02", "1970-01-01")

			newDate := t.AddDate(0, 0, int(i))
			p = fmt.Sprintf("'%s'", newDate.Format("2006-01-02"))
		case "io.debezium.time.Timestamp":
			i := v.(float64)
			nanos := int64(i) * 1000000
			var t time.Time

			if cfg.WithTimezone {
				t = time.Unix(0, nanos).Add(time.Hour * -7)
			} else {
				t = time.Unix(0, nanos)
			}

			p = fmt.Sprintf("'%s'", t.Format("2006-01-02 15:04:05"))
		case "io.debezium.time.MicroTime":
			i := v.(float64)
			d := time.Duration(i) * time.Microsecond
			second := d.Seconds()

			sec := int64(second) % 60
			min := int64(second) / 60 % 60
			hour := second / 3600
			hour = math.Floor(hour)

			p = fmt.Sprintf("'%d:%d:%d'", int64(hour), int64(min), int64(sec))

		case "io.debezium.time.ZonedTimestamp":
			p = v.(string)
			if re.MatchString(p) {
				p = strings.Replace(p, "T", " ", -1)
				p = strings.Replace(p, "Z", "", -1)
			}
			t, _ := time.Parse("2006-01-02 15:04:05", p)
			newDate := t.Add(time.Hour * 7)
			p = fmt.Sprintf("'%s'", newDate.Format("2006-01-02 15:04:05"))

		default:
			switch field.Type {
			case "boolean":
				ps := v.(bool)

				if ps {
					p = fmt.Sprintf("'%d'", 1)
				} else {
					p = fmt.Sprintf("'%d'", 0)
				}
			case "int64":
				ps := v.(float64)

				p = fmt.Sprintf("'%d'", int64(ps))
			case "float32", "float64":
				var final float64

				final, ok := v.(float64)
				rightVal := false

				if !ok {
					ps2, ok := v.(float32)
					if ok {
						rightVal = true
						final = float64(ps2)
					}
				} else {
					rightVal = true
				}

				if rightVal {
					p = fmt.Sprintf("'%f'", final)
				} else {
					p = fmt.Sprintf("'%v'", v)
				}

			default:
				p = strings.Replace(fmt.Sprintf("%v", v), "'", "", -1)

				p = fmt.Sprintf("'%s'", p)
			}
		}

		val = append(val, p)
		comb = append(comb, fmt.Sprintf("`%s` = %s ", k, p))
	}
	return strings.Join(key, ","), strings.Join(val, ","), strings.Join(comb, ","), strings.Join(comb, " AND ")
}

func processData(param []byte, cfg *config.Config) (string, error) {
	var expected data.Response

	err := json.Unmarshal(param, &expected)
	if err != nil {
		log.Println("parse err: ", string(param))
		return "", err
	}

	if len(expected.Payload.Source.Query) > 5 {
		return expected.Payload.Source.Query, nil
	}

	// processing
	tbl := expected.Payload.Source.Table

	qry := ""
	if expected.Payload.Before == nil && expected.Payload.After != nil {
		// insert query
		field, values, _, _ := mapToString(expected.Payload.After, expected.Schema.Fields, "after", cfg)
		qry = fmt.Sprintf("INSERT INTO `%s` (%s) VALUES (%s)", tbl, field, values)
	} else if expected.Payload.After != nil && expected.Payload.Before != nil {
		_, _, comb, _ := mapToString(expected.Payload.After, expected.Schema.Fields, "after", cfg)
		_, _, _, cond := mapToString(expected.Payload.Before, expected.Schema.Fields, "before", cfg)
		qry = fmt.Sprintf("UPDATE `%s` SET %s WHERE %s", tbl, comb, cond)
	} else if expected.Payload.Before != nil && expected.Payload.After == nil {
		// delete query
		_, _, _, cond := mapToString(expected.Payload.Before, expected.Schema.Fields, "before", cfg)
		qry = fmt.Sprintf("DELETE FROM `%s` WHERE %s", tbl, cond)
	}

	return qry, nil
}

func processScheme(param []byte, table []string, api *client.API, replaceAll bool, reclaim bool) (string, bool, error) {

	var expected scheme.Response
	err := json.Unmarshal(param, &expected)
	if err != nil {
		return "", false, err
	}

	if len(expected.Payload.DatabaseName) == 0 {
		// from instance
		return "", false, nil
	}

	if !replaceAll {
		found := false
		for _, v := range table {
			if expected.Payload.Source.Table == v {
				found = true
				break
			}
		}
		if !found {
			// is not source from this
			return "", false, nil
		}
	}

	if len(expected.Payload.DDL) == 0 {
		return "", false, fmt.Errorf("unexpected ddl")
	}

	if bl := utils.IsBlock(expected.Payload.DDL, reclaim); bl {
		return "", false, nil
	}

	// pause sync process
	err = api.Call("pause")
	if err != nil {
		return "", false, err
	}

	return expected.Payload.DDL, true, nil
}

func resume(api *client.API) {
	err := api.Call("resume")
	if err != nil {
		log.Println("failed stop sync ", err)
		time.Sleep(1 * time.Second)
		resume(api)
	}
}
