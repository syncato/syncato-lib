// Copyright 2015 The Syncato Authors.  All rights reserved.
// Use of this source code is governed by a AGPL
// license that can be found in the LICENSE file.

// Package auth defines the interface that authentication providers should implement and
// defines the authentication resource.
package auth

import (
	"fmt"
)

// AuthProvider is the interface that all the authentication providers must implement
// to be used by the authentication multiplexer.
// An authentication provider is defined by an ID.
// The extra parameter is useful to pass extra auth information to the underlying auth provider.
type AuthProvider interface {
	GetID() string
	Authenticate(username, password string, extra interface{}) (*AuthResource, error)
}

// AuthResource represents the details of an authenticated user.
type AuthResource struct {
	Username    string      `json:"username"`     // the ID for the user.
	DisplayName string      `json:"display_name"` // the user-friendly name.
	Email       string      `json:"email"`        // the email of the user.
	AuthID      string      `json:"auth_id"`      // the ID of the authentication provider who authenticated this user.
	Extra       interface{} `json:"extra"`
}

// UserNotFoundError represents a missing user in the authentication provider.
type UserNotFoundError struct {
	Username string
	AuthID   string
}

func (e *UserNotFoundError) Error() string {
	return fmt.Sprintf("user: %s not found in auth provider: %s", e.Username, e.AuthID)
}
