package config

import "github.com/kelseyhightower/envconfig"

// Config struct of configuration
type Config struct {
	// Kafka Information
	DBAddress    string `envconfig:"DBADDRESS" default:"ip_address_store_db"`
	DBSourceName string `envconfig:"DBSOURCE" default:"db_target"`
	DBUser       string `envconfig:"DBUSER" default:"db_user_targer"`
	DBPassword   string `envconfig:"DBPASSWORD" default:"db_password_target"`
	DBPort       int    `envconfig:"DBPORT" default:"3306"`
	DBLog        bool   `envconfig:"DB_LOG" default:"false"`
	Kafka        string `envconfig:"KAFKA" default:"localhost:9092"`
	Server       string `envconfig:"SERVER" default:"dbserver"`
	DBName       string `envconfig:"DBNAME" default:"db_name"`
	Table        string `envconfig:"TABLE" default:"table_name"`
	Group        string `envconfig:"GROUP" default:"name-group"`
	Republish    bool   `envconfig:"REPUBLISH" default:"false"`
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
