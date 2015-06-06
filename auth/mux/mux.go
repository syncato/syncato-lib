// Copyright 2015 The Syncato Authors.  All rights reserved.
// Use of this source code is governed by a AGPL
// license that can be found in the LICENSE file.

// Package mux defines the authentication multiplexer to authenticate requests against
// the registered authentication providers.
package mux

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/syncato/lib/auth"
	"github.com/syncato/lib/config"
	"github.com/syncato/lib/logger"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/net/context"
)

// AuthMux is the multiplexer responsible for routing authentication to an specific
// authentication provider.
// It keeps a map with all the authentication providers registered.
type AuthMux struct {
	cfg                     *config.Config
	log                     *logger.Logger
	registeredAuthProviders map[string]auth.AuthProvider
}

// NewAuthMux creates an AuthMux object or returns an error
func NewAuthMux(cfg *config.Config, log *logger.Logger) (*AuthMux, error) {
	m := AuthMux{}
	m.cfg = cfg
	m.log = log
	m.registeredAuthProviders = make(map[string]auth.AuthProvider)

	return &m, nil
}

// RegisterAuthProvider register an authentication providers to be used for authenticate requests.
func (mux *AuthMux) RegisterAuthProvider(ap auth.AuthProvider) error {
	if _, ok := mux.registeredAuthProviders[ap.GetID()]; ok {
		return errors.New(fmt.Sprintf("auth provider '%s' already registered", ap.GetID()))
	}
	mux.registeredAuthProviders[ap.GetID()] = ap
	return nil
}

// Authenticate authenticates a user with username and password credentials.
// The id parameter is the authentication provider id.
func (mux *AuthMux) Authenticate(username, password, id string, extra interface{}) (*auth.AuthResource, error) {
	// the authentication request has been made specifically for an authentication provider.
	if id != "" {
		a, ok := mux.registeredAuthProviders[id]
		// if an auth provider with the id passed is found we just use this auth provider.
		if ok {
			authRes, err := a.Authenticate(username, password, extra)
			if err != nil {
				return nil, err
			}
			return authRes, nil
		}
		return nil, &auth.UserNotFoundError{username, id}
	}

	// if the auth provider with the id passed is not found we try all the auth providers.
	// This is needed because with Basic Auth we cannot send the auth provider ID.
	for _, a := range mux.registeredAuthProviders {
		if a.GetID() != id {
			aRes, _ := a.Authenticate(username, password, extra)
			if aRes != nil {
				return aRes, nil
			}
		}
	}

	// we couldn´t find any auth provider that authenticated this user
	return nil, &auth.UserNotFoundError{username, "all"}
}

// AuthenticateRequest authenticates a HTTP request.
//
// It returns an AuthenticationResource object or an error.
//
// This method DOES NOT create an HTTP response with 401 if the authentication fails. To handle HTTP responses
// you must do it yourself or use the AuthMiddleware mehtod.
//
// The following mechanisms are used in the order described to authenticate the request.
//
// 1. JWT authentication token as query parameter in the URL. The parameter name is auth-key.
//
// 2. JWT authentication token in the HTTP Header called X-Auth-Key.
//
// 3. HTTP Basic Authentication without digest (Plain Basic Auth).
//
// More authentication methods wil be used in the future like Kerberos access tokens.
func (mux *AuthMux) AuthenticateRequest(r *http.Request) (*auth.AuthResource, error) {
	// 1. JWT authentication token as query parameter in the URL. The parameter name is auth-key.
	authQueryParam := r.URL.Query().Get("auth-key")
	if authQueryParam != "" {
		token, err := jwt.Parse(authQueryParam, func(token *jwt.Token) (key interface{}, err error) {
			return []byte(mux.cfg.TokenSecret()), nil
		})
		if err != nil {
			return nil, errors.New(fmt.Sprintf("Failed parsing auth query param because: %s", err.Error()))
		}
		authRes := &auth.AuthResource{}
		authRes.Username = token.Claims["username"].(string)
		authRes.DisplayName = token.Claims["display_name"].(string)
		authRes.Email = token.Claims["email"].(string)
		authRes.AuthID = token.Claims["auth_id"].(string)

		return authRes, nil
	}

	// 2. JWT authentication token in the HTTP Header called X-Auth-Key.
	authHeader := r.Header.Get("X-Auth-Key")
	if authHeader != "" {
		token, err := jwt.Parse(authHeader, func(token *jwt.Token) (key interface{}, err error) {
			return []byte(mux.cfg.TokenSecret()), nil
		})
		if err != nil {
			return nil, errors.New(fmt.Sprintf("failed parsing auth header because: %s", err.Error()))
		}
		authRes := &auth.AuthResource{}
		authRes.Username = token.Claims["username"].(string)
		authRes.DisplayName = token.Claims["display_name"].(string)
		authRes.Email = token.Claims["email"].(string)
		authRes.AuthID = token.Claims["auth_id"].(string)
		authRes.Extra = token.Claims["extra"]

		return authRes, nil
	}

	// 3. HTTP Basic Authentication without digest (Plain Basic Auth).
	username, password, ok := r.BasicAuth()
	if ok {
		authRes, err := mux.Authenticate(username, password, "", nil)
		if err != nil {
			return nil, err
		}
		if err == nil {
			return authRes, nil
		}
	}

	return nil, errors.New("no auth credentials found in the request")
}

// CreateAuthTokenFromAuthResource creates an JWT authentication token from an AuthenticationResource object.
// It returns the JWT token or an error.
func (mux *AuthMux) CreateAuthTokenFromAuthResource(authRes *auth.AuthResource) (string, error) {
	token := jwt.New(jwt.GetSigningMethod(mux.cfg.TokenCipherSuite()))
	token.Claims["iss"] = mux.cfg.TokenISS()
	token.Claims["exp"] = time.Now().Add(time.Minute * 480).Unix() // we need to use cfg.TokenExpirationTime
	token.Claims["username"] = authRes.Username
	token.Claims["display_name"] = authRes.DisplayName
	token.Claims["email"] = authRes.Email
	token.Claims["auth_id"] = authRes.AuthID

	tokenString, err := token.SignedString([]byte(mux.cfg.TokenSecret()))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

// AuthMiddleWare is an HTTP middleware that besides authenticating the request like the AuthenticateRequest method
// it does the following:
//
// 1. Return 401 (Unauthorized) if the authentication fails.
//
// 2. Save the AuthResource object in the request context and call the next handler if the authentication is successful.
func (mux *AuthMux) AuthMiddleware(ctx context.Context, w http.ResponseWriter, r *http.Request, next func(ctx context.Context, w http.ResponseWriter, r *http.Request)) {
	authRes, err := mux.AuthenticateRequest(r)
	if err != nil {
		mux.log.Error("Authentication of request failed", map[string]interface{}{"err": err})
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}
	mux.log.Info("Authentication of request successful", map[string]interface{}{"username": authRes.Username, "auth_id": authRes.AuthID})
	ctx = context.WithValue(ctx, "authRes", authRes)
	next(ctx, w, r)
}
