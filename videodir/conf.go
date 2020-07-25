package videodir

import (
	"github.com/BurntSushi/toml"
)

type Config struct {
	//LogLevel   string
	Debug      bool
	ServerAddr string
	Cert       string
	Key        string
	JwtSecret  string
	VideoDirs  []string

}

func DefaultConfig() Config {
	return Config{
		Debug:   	false,
		ServerAddr: "8443",
		Cert:       "server.crt",
		Key:        "server.key",
		JwtSecret:	"secret",
		VideoDirs:	[]string{},
	}
}

func (conf *Config) TOML(path string) error {
	_, err := toml.DecodeFile(path, &conf)
	return err
}
