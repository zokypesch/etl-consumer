# setup
# clone repo
git clone https://github.com/wurstmeister/kafka-docker
cd kafka-docker

# change docker-compose-single-broker.yml
```
version: '2'
services:
  zookeeper:
    image: wurstmeister/zookeeper
    ports:
      - "2181:2181"
  kafka:
    build: .
    ports:
      - "9092:9092"
    environment:
      KAFKA_ADVERTISED_HOST_NAME: 172.28.14.11 # your ip address
      KAFKA_ZOOKEEPER_CONNECT: zookeeper:2181
      KAFKA_CREATE_TOPICS: "my-topic:1:1"
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
```

# run single ??
docker-compose -f docker-compose-single-broker.yml up -d

# run dbzium 
docker run -it --rm --name connect -p 8083:8083 -e GROUP_ID=1 -e CONFIG_STORAGE_TOPIC=my_connect_configs -e OFFSET_STORAGE_TOPIC=my_connect_offsets -e STATUS_STORAGE_TOPIC=my_connect_statuses  -e BOOTSTRAP_SERVERS={kafka_address}:9092 -d debezium/connect:1.1

# create task
curl -i -X POST \
-H "Accept:application/json" \
-H "Content-Type:application/json" \
localhost:8083/connectors/ -d \
'{"name":"etl-connector-prod-region","config":{"connector.class":"io.debezium.connector.mysql.MySqlConnector","tasks.max":"100","database.hostname":"ip_address_database","database.port":"3306","database.user":"user_database","database.password":"password_database","database.server.id":"184088","database.server.name":"db_server_name","max.batch.size":32768,"max.queue.size":131072,"offset.flush.timeout.ms":60000,"offset.flush.interval.ms ":15000,"database.whitelist":"database-name","database.history.kafka.bootstrap.servers":"kafka_ip_address:9092","database.history.kafka.topic":"dbhistory.db_name"}}'

# delete task
curl -i -X DELETE \
-H "Accept:application/json" \
-H "Content-Type:application/json" \
localhost:8083/connectors/etl-connector-prod-region

# listen ???
# example response
```
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
		"seq": 0,
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
```

# how to install tresseract
brew install tesseract-lang
brew install imagemagick

sudo apt install tesseract-ocr
sudo apt install libtesseract-dev

# bugs
- handle decimal -> done
- handle date -> done
- remove resume and pause -> done
- publish topic to kafka (integration with engine)