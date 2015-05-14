package muxauth

import (
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/syncato/syncato-lib/auth"
	"github.com/syncato/syncato-lib/config"
	"github.com/syncato/syncato-lib/logger"
	"golang.org/x/net/context"
	"net/http"
	"strings"
	"time"
)

type MuxAuth struct {
	cp                      *config.ConfigProvider
	log                     *logger.Logger
	registeredAuthProviders map[string]auth.AuthProvider
}

func NewMuxAuth(cp *config.ConfigProvider, log *logger.Logger) (*MuxAuth, error) {
	m := MuxAuth{}
	m.cp = cp
	m.log = log
	m.registeredAuthProviders = make(map[string]auth.AuthProvider)

	return &m, nil
}

func (mux *MuxAuth) RegisterAuthProvider(ap auth.AuthProvider) error {
	if _, ok := mux.registeredAuthProviders[ap.GetID()]; ok {
		return &auth.AuthProviderAlreadyRegisteredError{ap.GetID()}
	}
	mux.registeredAuthProviders[ap.GetID()] = ap
	return nil
}

// Authenticate a user with username and password.
func (mux *MuxAuth) Authenticate(username, password, id string) (*auth.AuthResource, error) {
	a, ok := mux.registeredAuthProviders[id]
	if ok {
		authRes, err := a.Authenticate(username, password)
		if err != nil {
			return nil, err
		}
		return authRes, nil
	}
	for _, a := range mux.registeredAuthProviders {
		if a.GetID() != id {
			aRes, _ := a.Authenticate(username, password)
			if aRes != nil {
				return aRes, nil
			}
		}
	}
	return nil, &auth.UserNotFoundError{username, "all"}
}

// AuthenticateRequest authenticate a request.
// The credentials are checked in the following elements, in this order
// 1) Basic Auth
// 2) Auth token in header Authorization
// 3) Auth token in query param
func (mux *MuxAuth) AuthenticateRequest(r *http.Request) (*auth.AuthResource, error) {
	// Authenticate using basic auth
	username, password, ok := r.BasicAuth()
	if ok {
		authRes, err := mux.Authenticate(username, password, "")
		if err != nil {
			return nil, &auth.UserNotFoundError{username, "all"}
		}
		if err == nil {
			return authRes, nil
		}
	}

	cfg, err := mux.cp.ParseFile()
	if err != nil {
		return nil, err
	}

	if r.Header.Get("Authorization") != "" {
		tokenHeader := strings.Split(r.Header.Get("Authorization"), " ")
		if len(tokenHeader) >= 2 {
			token, err := jwt.Parse(string(tokenHeader[1]), func(token *jwt.Token) (key interface{}, err error) {
				return []byte(cfg.TokenSecret), nil
			})
			if err != nil {
				return nil, errors.New(fmt.Sprintf("Failed parsing auth token because: %s", err.Error()))
			}
			authRes := &auth.AuthResource{}
			authRes.Username = token.Claims["username"].(string)
			authRes.DisplayName = token.Claims["display_name"].(string)
			authRes.Email = token.Claims["email"].(string)
			authRes.AuthID = token.Claims["auth_id"].(string)

			return authRes, nil
		}
	}

	if r.URL.Query().Get("auth_token") != "" {
		token, err := jwt.Parse(string(r.URL.Query().Get("auth_token")), func(token *jwt.Token) (key interface{}, err error) {
			return []byte(cfg.TokenSecret), nil
		})
		if err != nil {
			return nil, errors.New(fmt.Sprintf("Failed parsing auth token because: %s", err.Error()))
		}
		authRes := &auth.AuthResource{}
		authRes.Username = token.Claims["username"].(string)
		authRes.DisplayName = token.Claims["display_name"].(string)
		authRes.Email = token.Claims["email"].(string)
		authRes.AuthID = token.Claims["auth_id"].(string)

		return authRes, nil
	}

	return nil, errors.New("No auth credentials found in the request")
}

func (mux *MuxAuth) CreateAuthTokenFromAuthResource(authRes *auth.AuthResource) (string, error) {
	cfg, err := mux.cp.ParseFile()
	if err != nil {
		return "", err
	}

	token := jwt.New(jwt.GetSigningMethod(cfg.TokenCipherSuite))
	token.Claims["iss"] = cfg.TokenISS
	token.Claims["exp"] = time.Now().Add(time.Minute * 480).Unix()
	token.Claims["username"] = authRes.Username
	token.Claims["display_name"] = authRes.DisplayName
	token.Claims["email"] = authRes.Email
	token.Claims["auth_id"] = authRes.AuthID

	tokenString, err := token.SignedString([]byte(cfg.TokenSecret))
	if err != nil {
		return "", nil
	}
	return tokenString, nil
}

func (mux *MuxAuth) AuthMiddleware(ctx context.Context, w http.ResponseWriter, r *http.Request, next func(ctx context.Context, w http.ResponseWriter, r *http.Request)) {
	authRes, err := mux.AuthenticateRequest(r)
	if err != nil {
		mux.log.Error("Authentication of request failed", map[string]interface{}{"err": err})
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	mux.log.Info("Authentication of request successful", map[string]interface{}{"username": authRes.Username, "auth_id": authRes.AuthID})
	ctx = context.WithValue(ctx, "authRes", authRes)
	next(ctx, w, r)
}
