// Copyright 2015 The Syncato Authors.  All rights reserved.
// Use of this source code is governed by a AGPL
// license that can be found in the LICENSE file.

// Package config defines the configuration provider to manipulate the daemon configuration file. Any custom fields
// added to the JSON configuration file should be declared here to be able to use them.
package config

import (
	"encoding/json"
	"github.com/syncato/lib/logger"
	"io/ioutil"
	"os"
	"sync"
)

// ConfigParams represents the structure of the configuration file used by the daemon.
// This is a sample JSON configuration file:
// 	{
// 	  "installed": true,
// 	  "maintenance": false,
// 	  "maintenance_message": "Sorry for the intervention",
// 	  "token_secret": "ViYiQqErJrt045NoRxCO0f3S6yzOuqRBZF8GyanMhwl1RyGB9GGe2KlGe3XR",
// 	  "token_cipher_suite": "HS256",
// 	  "token_iss": "syncato.org",
// 	  "token_expiration_time": 3600,
// 	  "create_user_home_on_login": true,
// 	  "create_user_home_in_storages": ["local"]
// 	  "root_data_dir": "/data",
// 	  "root_tmp_dir": "/tmp",
// 	  "auth_json_file": "/etc/private/syncato_auth.json"
// 	}
type ConfigParams struct {

	// @RW
	// Indicates if the daemon has been installed.
	Installed bool `json:"installed"`

	// @RW
	// Indicates if the daemon is in maintenance mode.
	// All the responses will be 503 (Temporary Unavailable).
	Maintenance bool `json:"maintenance"`

	// @RW
	// If the daemon is in maintenance mode, indicates a custom message to serve.
	// If this is empty, the default message will be "Temporary unavailable".
	MaintenanceMessage string `json:maintenance_message`

	// @RO
	// The JSON web token secret used to encrypt sensitive data.
	// Once the daemon has run you MUST NOT change this value.
	// Extended documentation about JSON Web Tokens (JWT) can be found
	// at http://self-issued.info/docs/draft-ietf-oauth-json-web-token.html
	TokenSecret string `json:"token_secret"`

	// @RO
	// The cipher suite used to create the JWT secret.
	// Once the daemon has run you MUST NOT change this value.
	// Possible values: HS256
	TokenCipherSuite string `json:"token_cipher_suite"`

	// @RO
	// The name of the organization issuing the JWT.
	TokenISS string `json:"token_iss"`

	// @RO
	// The duration in seconds of the JWT to be valid.
	TokenExpirationTime int `json:"token_expiration_time"`

	// @RW
	// Indicates if the user homedirectory must be created when the user log in
	CreateUserHomeOnLogin bool `json:"create_user_home_on_login"`

	// @RW
	// If CreateUserHomeOnLogin is enabled indicates in which storages the home dir will be created.
	CreateUserHomeInStorages []string `json:"create_user_home_in_storages"`

	// @RW
	// Indicates if the server should validate the upload with the provided checksum and checksumtype
	// sent by the client.
	// If the checksum type sended by the client is not supported by the server the upload will fail.
	// If this option is enabled and the checksum type is empty, checksum validation will not be triggered.
	VerifyClientChecksum bool `json:"verify_client_checksum"`

	// @RW
	// Indicates if the server shoudl send the X-Checksum header with the checksum of the file the client
	// is asking to download.
	// To use this feature the client must send a query parameter called checksumtype that specifies the
	// checksum the server should send.
	// It is up to the client to validate the download of the file against his header.
	SendChecksumHeader bool `json:"send_checksum_header"`

	// @RO
	// Indicates where data will be saved.
	RootDataDir string `json:"root_data_dir"`

	// @RO
	// Indicates where temporary data will be saved.
	RootTmpDir string `json:"root_tmp_dir"`

	// @RO
	// Indicates the JSON file to be used as an authentication backend.
	AuthJSONFile string `json:"auth_json_file"`
}

func New(filename string, log *logger.Logger) (*Config, error) {
	var cfg = &ConfigParams{}
	fd, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	data, err := ioutil.ReadAll(fd)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(data, cfg)
	if err != nil {
		return nil, err
	}
	rcfg := &Config{
		filename: filename,
		cfg:      cfg,
	}
	return rcfg, nil
}
func NewWithModel(filename string, cfg *ConfigParams, log *logger.Logger) (*Config, error) {
	fd, err := os.Create(filename + ".tmp")
	if err != nil {
		return nil, err
	}
	data, err := json.Marshal(cfg)
	if err != nil {
		return nil, err
	}
	_, err = fd.Write(data)
	if err != nil {
		return nil, err
	}
	return New(filename, log)
}

type Config struct {
	filename string
	cfg      *ConfigParams
	sync.Mutex
	log *logger.Logger
}

func (c *Config) save() error {
	fd, err := os.Create(c.filename + ".tmp")
	if err != nil {
		return err
	}
	data, err := json.Marshal(c.cfg)
	if err != nil {
		return err
	}
	_, err = fd.Write(data)
	if err != nil {
		return err
	}
	return os.Rename(c.filename+".tmp", c.filename)
}
func (c *Config) Reload() error {
	var cfg = &ConfigParams{}
	fd, err := os.Open(c.filename)
	if err != nil {
		return err
	}
	data, err := ioutil.ReadAll(fd)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, cfg)
	if err != nil {
		return err
	}
	c.Lock()
	c.cfg = cfg
	c.Unlock()
	return nil
}

func (c *Config) Maintenance() bool {
	return c.cfg.Maintenance
}
func (c *Config) SetMaintenance(val bool) error {
	c.Lock()
	c.cfg.Maintenance = val
	err := c.save()
	c.Unlock()
	return err
}
func (c *Config) MaintenanceMessage() string {
	return c.cfg.MaintenanceMessage
}
func (c *Config) SetMaintenanceMessage(msg string) error {
	c.Lock()
	c.cfg.MaintenanceMessage = msg
	err := c.save()
	c.Unlock()
	return err
}
func (c *Config) TokenSecret() string {
	return c.cfg.TokenSecret
}
func (c *Config) TokenCipherSuite() string {
	return c.cfg.TokenCipherSuite
}
func (c *Config) TokenISS() string {
	return c.cfg.TokenISS
}
func (c *Config) CreateUserHomeOnLogin() bool {
	return c.cfg.CreateUserHomeOnLogin
}
func (c *Config) SetCreateUserHomeOnLogin(val bool) error {
	c.Lock()
	c.cfg.CreateUserHomeOnLogin = val
	err := c.save()
	c.Unlock()
	return err
}
func (c *Config) CreateUserHomeInStorages() []string {
	return c.cfg.CreateUserHomeInStorages
}
func (c *Config) SetCreateUserHomeInStorages(val []string) error {
	c.Lock()
	c.cfg.CreateUserHomeInStorages = val
	err := c.save()
	c.Unlock()
	return err
}
func (c *Config) VerifyClientChecksum() bool {
	return c.cfg.VerifyClientChecksum
}
func (c *Config) SetVerifyClientChecksum(val bool) error {
	c.Lock()
	c.cfg.VerifyClientChecksum = val
	err := c.save()
	c.Unlock()
	return err
}
func (c *Config) SendChecksumHeader() bool {
	return c.cfg.SendChecksumHeader
}
func (c *Config) SetSendChecksumHeader(val bool) error {
	c.Lock()
	c.cfg.SendChecksumHeader = val
	err := c.save()
	c.Unlock()
	return err
}
func (c *Config) RootDataDir() string {
	return c.cfg.RootDataDir
}
func (c *Config) RootTmpDir() string {
	return c.cfg.RootTmpDir
}
func (c *Config) AuthJSONFile() string {
	return c.cfg.AuthJSONFile
}
