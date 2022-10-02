package main

type Config struct {
	Token            string `json:"token"`
	Https            bool   `json:"https"`
	DaemonHost       string `json:"daemonHost"`
	APIHost          string `json:"apiHost"`
	DataHost         string `json:"dataHost"`
	ReconnectTimeout uint16 `json:"reconnectTimeout"`
}

func NewConfig() Config {
	return Config{
		Https:            false,
		DaemonHost:       "localhost:9002",
		APIHost:          "localhost:9001",
		DataHost:         "localhost:455",
		ReconnectTimeout: 5,
	}
}

type ConfigCredentials struct {
	Databases []ConfigCredentialsDatabase
	Smb       []ConfigCredentialsSmbUser
}

type ConfigCredentialsDatabase struct {
	Host      string                              `json:"host"`
	Port      string                              `json:"port"`
	Username  string                              `json:"username"`
	Password  string                              `json:"password"`
	Databases []ConfigCredentialsDatabaseDatabase `json:"databases"`
}

type ConfigCredentialsDatabaseDatabase struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type ConfigCredentialsSmbUser struct {
	ID       string `json:"id"`
	Password string `json:"password"`
}

func NewConfigCredentials() ConfigCredentials {
	return ConfigCredentials{
		Databases: []ConfigCredentialsDatabase{},
		Smb:       []ConfigCredentialsSmbUser{},
	}
}
