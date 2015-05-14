package auth

import (
	"fmt"
)

// AuthProvider is the interface that all the authentication providers must implement
// Authenticate authenticate the request and return an authentication resource
// or a validation error or a server error
type AuthProvider interface {
	GetID() string
	Authenticate(username, password string) (*AuthResource, error)
}

// AuthResource represents the user after a valid authentication
// Username is the primary identifier
// DisplayName is the user friendly name for the user id
// Email is the email of the user
// Auth is the type of authentication used for this user
type AuthResource struct {
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	Email       string `json:"email"`
	AuthID      string `json:"auth_id"`
}

/*
func GetUserCredentialsFromBasicAuth(r *http.Request) (username, password string) {
	username, password, ok := r.BasicAuth()
	if !ok {
		username, password = "", ""
	}
	return username, password
}
*/

type UserNotFoundError struct {
	Username string
	AuthID   string
}

func (e *UserNotFoundError) Error() string {
	return fmt.Sprintf("user: %s not found in auth provider: %s", e.Username, e.AuthID)
}

type AuthProviderAlreadyRegisteredError struct {
	AuthProviderID string
}

func (e *AuthProviderAlreadyRegisteredError) Error() string {
	return fmt.Sprintf("auth provider:%s already registred", e.AuthProviderID)
}

type AuthProviderNotRegisteredError struct {
	AuthProviderID string
}

func (e *AuthProviderNotRegisteredError) Error() string {
	return fmt.Sprintf("auth provider:%s is not registered", e.AuthProviderID)
}
