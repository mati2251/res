package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	BaseDir string `yaml:"base_dir"`
	Port    int    `yaml:"port"`
	DbUrl   string `yaml:"db_url"`
}

var config Config

func Get(path string) (Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return Config{}, err
	}
	defer file.Close()
	var config Config
	err = yaml.NewDecoder(file).Decode(&config)
	if err != nil {
		return Config{}, err
	}
	return config, nil
}
