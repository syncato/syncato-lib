package config

import (
	"encoding/json"
	"github.com/syncato/syncato-lib/logger"
	"io/ioutil"
	"os"
)

type Config struct {

	// SERVER CONFIG
	Installed   bool `json:"installed"`
	Maintenance bool `json:"maintenance"`
	Port        int  `json:"port"`

	// TOKEN CONFIGURATION
	TokenSecret      string `json:"token_secret"`
	TokenCipherSuite string `json:"token_cipher_suite"`
	TokenISS         string `json:"token_iss"`

	// WEB CONFIG
	ServeWeb string `json:"serve_web"`
	WebDir   string `json:"web_dir"`
	WebURL   string `json:"web_url"`

	// LOCAL STORAGE OPTIONS
	RootDataDir string `json:"root_data_dir"`
	RootTmpDir  string `json:"root_tmp_dir"`

	// JSON AUTH OPTIONS
	JSONAuthFile string `json:"json_auth_file"`
}

type ConfigProvider struct {
	configFilename string
	log            *logger.Logger
}

func NewConfigProvider(configFilename string, log *logger.Logger) (*ConfigProvider, error) {
	return &ConfigProvider{configFilename, log}, nil
}

func (cp *ConfigProvider) ParseFile() (*Config, error) {
	fd, err := os.Open(cp.configFilename)
	if err != nil {
		return nil, err
	}
	data, err := ioutil.ReadAll(fd)
	if err != nil {
		return nil, err
	}
	var config Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}
func (cp *ConfigProvider) CreateConfig(cfg *Config, configFilename string) error {
	fd, err := os.Create(configFilename)
	if err != nil {
		return err
	}
	data, err := json.Marshal(cfg)
	if err != nil {
		return err
	}
	_, err = fd.Write(data)
	if err != nil {
		return err
	}
	return nil
}
