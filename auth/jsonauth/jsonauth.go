package jsonauth

import (
	"encoding/json"
	"github.com/syncato/syncato-lib/auth"
	"github.com/syncato/syncato-lib/config"
	"github.com/syncato/syncato-lib/logger"
	"io/ioutil"
	"os"
)

// User reprents a user saved in the json auth file
type User struct {
	Username    string `json:"username"`
	Password    string `json:"password"`
	DisplayName string `json:"display_name"`
	Email       string `json:"email"`
}

type JSONAuth struct {
	id  string
	cp  *config.ConfigProvider
	log *logger.Logger
}

func NewJSONAuth(id string, cp *config.ConfigProvider, log *logger.Logger) (*JSONAuth, error) {
	return &JSONAuth{id, cp, log}, nil
}

func (a *JSONAuth) GetID() string {
	return a.id
}

func (a *JSONAuth) Authenticate(username, password string) (*auth.AuthResource, error) {
	cfg, err := a.cp.ParseFile()
	if err != nil {
		a.log.Error(err.Error(), nil)
		return nil, err
	}
	fd, err := os.Open(cfg.JSONAuthFile)
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
			}
			return &authRes, nil
		}
	}
	return nil, &auth.UserNotFoundError{username, a.GetID()}
}

/*
func (a *JSONAuth) CreateUser(user *User) error {
	users := []*User{user}
	fd, err := os.Create(a.jsonAuthFile)
	if err != nil {
		return err
	}
	data, err := json.Marshal(users)
	if err != nil {
		return err
	}
	_, err = fd.Write(data)
	if err != nil {
		return err
	}
	return nil
}
*/
