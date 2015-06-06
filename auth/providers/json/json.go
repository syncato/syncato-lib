// Copyright 2015 The Syncato Authors.  All rights reserved.
// Use of this source code is governed by a AGPL
// license that can be found in the LICENSE file.

// Package json implements the AuthProvider interface to authenticate users agains a JSON file.
package json

import (
	"encoding/json"
	"github.com/syncato/lib/auth"
	"github.com/syncato/lib/config"
	"github.com/syncato/lib/logger"
	"io/ioutil"
	"os"
)

// User reprents a user saved in the JSON authentication file.
type User struct {
	Username    string      `json:"username"`
	Password    string      `json:"password"`
	DisplayName string      `json:"display_name"`
	Email       string      `json:"email"`
	Extra       interface{} `json:"extra"`
}

// AuthJSON is the implementation of the AuthProvider interface to use a JSON
// file as an autentication provider.
// This authentication provider should be used just for testing or for small installations.
type AuthJSON struct {
	id  string
	cfg *config.Config
	log *logger.Logger
}

// NewAuthJSON returns an AuthJSON object or an error.
func NewAuthJSON(id string, cfg *config.Config, log *logger.Logger) (*AuthJSON, error) {
	return &AuthJSON{id, cfg, log}, nil
}

// GetID returns the ID of the JSON auth provider.
func (a *AuthJSON) GetID() string {
	return a.id
}

// Authenticate authenticates a user agains the JSON file.
// User credentials in the JSON file are kept in plain text, so the password is not encrypted.
func (a *AuthJSON) Authenticate(username, password string, extra interface{}) (*auth.AuthResource, error) {
	fd, err := os.Open(a.cfg.AuthJSONFile())
	defer fd.Close()
	if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadAll(fd)
	if err != nil {
		a.log.Error(err.Error(), nil)
		return nil, err
	}

	users := make([]*User, 0)
	err = json.Unmarshal(data, &users)
	if err != nil {
		a.log.Error(err.Error(), nil)
		return nil, err
	}

	for _, user := range users {
		if user.Username == username && user.Password == password {
			authRes := auth.AuthResource{
				Username:    user.Username,
				DisplayName: user.DisplayName,
				Email:       user.Email,
				AuthID:      a.GetID(),
				Extra:       user.Extra,
			}
			return &authRes, nil
		}
	}
	return nil, &auth.UserNotFoundError{username, a.GetID()}
}
