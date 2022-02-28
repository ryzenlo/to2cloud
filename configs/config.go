package configs

import (
	"io/ioutil"
	"log"

	env "github.com/Netflix/go-env"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Mongo   MongoConfig   `yaml:"mongo"`
	MySQL   MySQLConfig   `yaml:"mysql"`
	Sqlite  SqliteConfig  `yaml:"sqlite"`
	Redis   RedisConfig   `yaml:"redis"`
	JWT     JWTConfig     `yaml:"jwt"`
	Log     LogConfig     `yaml:"log"`
	Ansible AnsibleConfig `yaml:"ansible"`
}

type LogConfig struct {
	Level    string `yaml:"level"`
	DirPath  string `yaml:"dir_path"`
	FileName string `yaml:"file_name"`
	MaxSize  int64  `yaml:"max_size"`
}

type SqliteConfig struct {
	DirPath    string `yaml:"dir_path"`
	DBFileName string `yaml:"db_filename"`
}

type MySQLConfig struct {
}

type MongoConfig struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
}

type RedisConfig struct {
}

type JWTConfig struct {
	Key string `yaml:"key" env:"JWT_KEY"`
}

type AnsibleConfig struct {
	DirPath string `yaml:"dir_path"`
}

type ProxyConfig struct {
	UseProxy bool   `yaml:"use_proxy" json:"use_proxy"`
	Type     string `yaml:"proxy_type" json:"proxy_type"`
	Host     string `yaml:"proxy_host" json:"proxy_host"`
	Port     string `yaml:"proxy_port" json:"proxy_port"`
}

var Conf *Config

func init() {
	Conf = &Config{}
}

func LoadConfigFile(filePath string) {
	bs, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Fatalf("Load config file failed, %v", err)
	}
	err = yaml.Unmarshal(bs, Conf)
	if err != nil {
		log.Fatalf("Load config file failed, %v", err)
	}
	if _, err := env.UnmarshalFromEnviron(Conf); err != nil {
		log.Fatalf("Load config file failed, %v", err)
	}
}
