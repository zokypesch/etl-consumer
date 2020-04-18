package config

import "github.com/kelseyhightower/envconfig"

// Config struct of configuration
type Config struct {
	// Kafka Information
	Kafka  string `envconfig:"KAFKA" default:"localhost:9092"`
	Server string `envconfig:"SERVER" default:"server_name"`
	DBName string `envconfig:"DBNAME" default:"db_name"`
	Table  string `envconfig:"TABLE" default:"table_name"`
	Group  string `envconfig:"GROUP" default:"group-name"`
	// FB Information
	DBAddress    string `envconfig:"DBADDRESS" default:"localhost"`
	DBSourceName string `envconfig:"DBNAME" default:"etl"`
	DBUser       string `envconfig:"DBUSER" default:"root"`
	DBPassword   string `envconfig:"DBPASSWORD" default:""`
	DBPort       int    `envconfig:"DBPORT" default:"3306"`
	DBLog        bool   `envconfig:"DB_LOG" default:"false"`
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
