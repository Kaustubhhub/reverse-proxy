package config

type Config struct {
	Servers []string `yaml:"servers"`
}

const SERVERS_PATH = "./config/servers.yaml"
