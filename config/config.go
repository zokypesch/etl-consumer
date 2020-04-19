package config

import "github.com/kelseyhightower/envconfig"

// Config struct of configuration
type Config struct {
	// Kafka Information
	DBAddress    string `envconfig:"DBADDRESS" default:"localhost"`
	DBSourceName string `envconfig:"DBSOURCE" default:"master_etl"`
	DBUser       string `envconfig:"DBUSER" default:"etl_master"`
	DBPassword   string `envconfig:"DBPASSWORD" default:"password"`
	DBPort       int    `envconfig:"DBPORT" default:"3306"`
	DBLog        bool   `envconfig:"DB_LOG" default:"false"`
	Kafka        string `envconfig:"KAFKA" default:"localhost:9092"`
	Server       string `envconfig:"SERVER" default:"server_debezium"`
	DBName       string `envconfig:"DBNAME" default:"your_db_debezium"`
	Table        string `envconfig:"TABLE" default:"table_name"`
	Group        string `envconfig:"GROUP" default:"name-group"`
	ActiveScheme bool   `envconfig:"ACTIVE_SCHEME" default:"false"`
	Republish    bool   `envconfig:"REPUBLISH" default:"true"`
	Connector    string `envconfig:"CONNECTOR" default:"connector-name"`
	DebeziumAddr string `envconfig:"DEBEZIUM_ADDR" default:"connector-address"` //localhost
	DebeziumPort string `envconfig:"DEBEZIUM_PORT" default:"8083"`
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
