package config

import (
	"gochat/core"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
)

type Server struct {
	Network string
	Address string
}

type Client struct {
	Network string
	Address string
}

type Config struct {
	Server Server
	Client Client
	Users  []core.User
}

var config *Config // global app config

func GetConfig() *Config {
	return config
}

func LoadConfig() {
	file, err := os.Open("conf.yaml")
	if err != nil {
		panic(err)
	}

	bytes, err := ioutil.ReadAll(
		file)
	if err != nil {
		panic(err)
	}

	cfg := Config{}
	err = yaml.Unmarshal(bytes, &cfg)
	if err != nil {
		panic(err)
	}

	config = &cfg
}
