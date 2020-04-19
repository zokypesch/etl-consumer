package config

import "github.com/kelseyhightower/envconfig"

// Config struct of configuration
type Config struct {
	// Kafka Information
	DBAddress    string `envconfig:"DBADDRESS" default:"rm-d9j49704dut1q2s9v.mysql.ap-southeast-5.rds.aliyuncs.com"`
	DBSourceName string `envconfig:"DBSOURCE" default:"master_etl"`
	DBUser       string `envconfig:"DBUSER" default:"etl_master"`
	DBPassword   string `envconfig:"DBPASSWORD" default:"Pr4K3rj4S3laM4ny4"`
	DBPort       int    `envconfig:"DBPORT" default:"3306"`
	DBLog        bool   `envconfig:"DB_LOG" default:"false"`
	Kafka        string `envconfig:"KAFKA" default:"172.28.14.13:9092"`

	Server           string `envconfig:"SERVER" default:"dbserver_alibaba"`
	DBName           string `envconfig:"DBNAME" default:"alibaba"`
	Table            string `envconfig:"TABLE" default:"group"`
	Group            string `envconfig:"GROUP" default:"alibaba-group"`
	AutoOffset       string `envconfig:"AUTO_OFFSET" default:"latest"` // earliest
	ActiveScheme     bool   `envconfig:"ACTIVE_SCHEME" default:"true"`
	ReplaceAllScheme bool   `envconfig:"REPLACE_ALL_SCHEME" default:"true"`
	Republish        bool   `envconfig:"REPUBLISH" default:"false"`
	Connector        string `envconfig:"CONNECTOR" default:"etl-connector-prod-alibaba"`
	DebeziumAddr     string `envconfig:"DEBEZIUM_ADDR" default:"172.28.14.13"` //localhost
	DebeziumPort     string `envconfig:"DEBEZIUM_PORT" default:"8083"`
}

// singleton of data
var data *Config

// Get configuration of data
func Get() *Config {
	if data == nil {
		data = &Config{}
		envconfig.MustProcess("", data)
	}

	// returing configuration
	return data
}
