package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	BaseDir string `yaml:"base_dir"`
	Port    int    `yaml:"port"`
}

var config Config

func get(path string) (Config, error) {
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

func Set(path string) (Config, error) {
	conf, err := get(path)
	if err != nil {
		return Config{}, err
	}
	config = conf
	return config, nil
}

func Get() Config {
	return config
}
