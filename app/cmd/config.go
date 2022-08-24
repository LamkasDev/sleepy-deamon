package main

type Config struct {
	Token               string                      `json:"token"`
	Https               bool                        `json:"https"`
	DaemonHost          string                      `json:"daemonHost"`
	APIHost             string                      `json:"apiHost"`
	DataHost            string                      `json:"dataHost"`
	ReconnectTimeout    uint16                      `json:"reconnectTimeout"`
	DatabaseCredentials []ConfigDatabaseCredentials `json:"databaseCredentials"`
}

type ConfigDatabaseCredentials struct {
	Host      string   `json:"host"`
	Port      string   `json:"port"`
	Username  string   `json:"username"`
	Password  string   `json:"password"`
	Databases []string `json:"databases"`
}

func NewConfig() Config {
	return Config{
		Https:               false,
		DaemonHost:          "localhost:9002",
		APIHost:             "localhost:9001",
		DataHost:            "localhost:455",
		ReconnectTimeout:    5,
		DatabaseCredentials: []ConfigDatabaseCredentials{},
	}
}
