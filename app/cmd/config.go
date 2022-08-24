package main

type Config struct {
	Token               string                      `json:"token"`
	Https               bool                        `json:"https"`
	Host                string                      `json:"host"`
	APIPort             uint16                      `json:"apiPort"`
	StaticPort          uint16                      `json:"staticPort"`
	Port                uint16                      `json:"port"`
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
		Host:                "localhost",
		APIPort:             8081,
		StaticPort:          8082,
		Port:                23365,
		ReconnectTimeout:    5,
		DatabaseCredentials: []ConfigDatabaseCredentials{},
	}
}
