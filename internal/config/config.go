package config

import (
	"fmt"
	"os"
	"strconv"

	"gopkg.in/yaml.v3"
)

var loaded Config

type Config struct {

	Server struct {
		Port uint16 `yaml:"port"`
	} `yaml:"server"`

	Database struct {
		Host			string `yaml:"host"`
		Port			uint16 `yaml:"port"`
		Name			string `yaml:"name"`
		User			string `yaml:"user"`
		Password	string `yaml:"password"`
	} `yaml:"database"`
	
	SSH struct {
		User			string `yaml:"user"`
		KeyPath		string `yaml:"key_path"`
		Port			uint8	 `yaml:"port"`
		Timeout		uint8  `yaml:"timeout_seconds"`
	} `yaml:"ssh"`

	Flags struct {
		MaxPlayers							uint32	`yaml:"max_players"`
		MinPort									uint32	`yaml:"min_port"`
		MaxPort									uint32	`yaml:"max_port"`
		HostCheckInterval				uint16	`yaml:"host_check_interval_seconds"`
		InstanceCheckInterval		uint16	`yaml:"instance_check_interval_seconds"`
		MinRAM									uint32	`yaml:"min_ram"`
		MaxRAM									uint32	`yaml:"max_ram"`
	} `yaml:"flags"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err = yaml.Unmarshal([]byte(os.ExpandEnv(string(data))), &cfg); err != nil {
		return nil, err
	}
	
	loaded = cfg

	return &cfg, nil
}

func GetDatabaseURL(cfg *Config) (string, error) {
	if cfg == nil {
		return "", fmt.Errorf("cfg cannot be nil")
	}
	
	port := strconv.FormatUint(uint64(cfg.Database.Port), 10)
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", cfg.Database.User, cfg.Database.Password, cfg.Database.Host, port, cfg.Database.Name), nil
}

func ReadOnlyInstance() Config {
	return loaded
}
