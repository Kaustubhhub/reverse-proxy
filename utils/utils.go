package utils

import (
	"os"

	"github.com/kaustubhhub/reverse-proxy/config"
	"gopkg.in/yaml.v3"
)

func GetNextBackendServer(backendServerList []string, currentBackendServer *int) int {
	*currentBackendServer = (*currentBackendServer + 1) % len(backendServerList)
	return *currentBackendServer
}

func LoadConfig(path string) (*config.Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg config.Config

	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}
