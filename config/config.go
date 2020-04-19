package config

import "github.com/kelseyhightower/envconfig"

// Config struct of configuration
type Config struct {
	// Kafka Information
	DBAddress        string   `envconfig:"DBADDRESS" default:"localhost"`
	DBSourceName     string   `envconfig:"DBSOURCE" default:"master_etl"`
	DBUser           string   `envconfig:"DBUSER" default:"etl_master"`
	DBPassword       string   `envconfig:"DBPASSWORD" default:"your_db_password"`
	DBPort           int      `envconfig:"DBPORT" default:"3306"`
	DBLog            bool     `envconfig:"DB_LOG" default:"false"`
	Kafka            string   `envconfig:"KAFKA" default:"172.28.14.13:9092"`
	Reclaim          bool     `envconfig:"RECLAIM" default:"false"`
	Server           string   `envconfig:"SERVER" default:"dbserver_name"`
	DBName           string   `envconfig:"DBNAME" default:"db_name"`
	Table            []string `envconfig:"TABLE" default:"table_name1,table2"`
	Group            string   `envconfig:"GROUP" default:"name-group"`
	Republish        bool     `envconfig:"REPUBLISH" default:"true"`
	RepublishLimit   int      `envconfig:"REPUBLISH_LIMIT" default:"3"`
	Connector        string   `envconfig:"CONNECTOR" default:"etl-connector-name"`
	DebeziumAddr     string   `envconfig:"DEBEZIUM_ADDR" default:"172.28.14.13"`
	DebeziumPort     string   `envconfig:"DEBEZIUM_PORT" default:"8083"`
	AutoOffset       string   `envconfig:"AUTO_OFFSET" default:"latest"` // earliest
	ActiveScheme     bool     `envconfig:"ACTIVE_SCHEME" default:"false"`
	ReplaceAllScheme bool     `envconfig:"REPLACE_ALL_SCHEME" default:"true"`
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
