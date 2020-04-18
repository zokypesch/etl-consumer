package main

import (
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/zokypesch/etl/config"
	"github.com/zokypesch/etl/data"
	"github.com/zokypesch/proto-lib/core"
	"gopkg.in/confluentinc/confluent-kafka-go.v1/kafka"
)

var testData = `
{
	"payload": {
		"before": {
		"id": 50,
		"province_id": 999,
		"seq": 0,
		"created_at": "2020-04-18T08:54:10Z"
		},
		"after": {
		"id": 50,
		"province_id": 999,
		"seq": null,
		"created_at": "2020-04-18T08:54:10Z"
		},
		"source": {
		"version": "1.1.1.Final",
		"connector": "mysql",
		"name": "dbserver2",
		"ts_ms": 0,
		"snapshot": "true",
		"db": "batch",
		"table": "batch_seq",
		"server_id": 0,
		"gtid": null,
		"file": "mysql-bin.000082",
		"pos": 289021,
		"row": 0,
		"thread": null,
		"query": null
		},
		"op": "c",
		"ts_ms": 1587202401764,
		"transaction": null
	}
}
`

func main() {

	log.Fatal(processData([]byte(testData)))

	cfg := config.Get()
	db := core.InitDB(cfg.DBAddress, cfg.DBSourceName, cfg.DBUser, cfg.DBPassword, cfg.DBPort, cfg.DBLog)

	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": cfg.Kafka,
		"group.id":          cfg.Group,
		"auto.offset.reset": "earliest",
	})

	if err != nil {
		panic(err)
	}

	// republish message
	p, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": cfg.Kafka})

	if err != nil {
		panic(err)
	}

	topicName := fmt.Sprintf("%s.%s.%s", cfg.Server, cfg.DBName, cfg.Table)
	c.SubscribeTopics([]string{topicName}, nil)
	log.Printf("starting subscribe %s.%s.%s", cfg.Server, cfg.DBName, cfg.Table)
	for {
		msg, err := c.ReadMessage(-1)
		if err == nil {
			log.Printf("Execute message on %s: %s\n", msg.TopicPartition, string(msg.Value))
			qry := processData(msg.Value)
			err = db.Exec(qry).Error
			if err != nil {
				log.Fatal("error exec qry ", err)
				db.Exec(fmt.Sprintf("INSERT INTO data_err (data, error) VALUES('%s', '%s')", string(msg.Value), err.Error()))
				if cfg.Republish {
					deliveryChan := make(chan kafka.Event)

					err = p.Produce(&kafka.Message{
						TopicPartition: kafka.TopicPartition{Topic: &topicName, Partition: kafka.PartitionAny},
						Value:          []byte(msg.Value),
						// Headers:        []kafka.Header{{Key: "myTestHeader", Value: []byte("header values are binary")}},
					}, deliveryChan)

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
			}

		} else {
			// The client will automatically try to recover from all errors.
			log.Printf("Consumer error: %v (%v)\n", err, msg)
		}
	}

	c.Close()
}

func mapToString(param map[string]interface{}) (string, string, string, string) {
	var key []string
	var val []string
	var comb []string

	re := regexp.MustCompile("((19|20)\\d\\d)-(0?[1-9]|1[012])-(0?[1-9]|[12][0-9]|3[01])")

	for k, v := range param {
		if v == nil {
			continue
		}
		key = append(key, k)

		p := fmt.Sprintf("'%v'", v)
		if re.MatchString(p) {
			p = strings.Replace(p, "T", " ", -1)
			p = strings.Replace(p, "Z", "", -1)
		}
		val = append(val, p)
		comb = append(comb, fmt.Sprintf("%s = %s ", k, p))
	}
	return strings.Join(key, ","), strings.Join(val, ","), strings.Join(comb, ","), strings.Join(comb, " AND ")
}

func processData(param []byte) string {
	var expected data.Response

	err := json.Unmarshal(param, &expected)
	if err != nil {
		log.Println(err)
	}

	if len(expected.Payload.Source.Query) > 5 {
		return expected.Payload.Source.Query
	}

	// processing
	tbl := expected.Payload.Source.Table

	qry := ""
	if expected.Payload.Before == nil && expected.Payload.After != nil {
		// insert query
		field, values, _, _ := mapToString(expected.Payload.After)
		qry = fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", tbl, field, values)
	} else if expected.Payload.After != nil && expected.Payload.Before != nil {
		_, _, comb, _ := mapToString(expected.Payload.After)
		_, _, _, cond := mapToString(expected.Payload.Before)
		qry = fmt.Sprintf("UPDATE %s SET %s WHERE %s", tbl, comb, cond)
	} else if expected.Payload.Before != nil && expected.Payload.After == nil {
		// delete query
		_, _, _, cond := mapToString(expected.Payload.Before)
		qry = fmt.Sprintf("DELETE FROM %s WHERE %s", tbl, cond)
	}

	return qry
}
