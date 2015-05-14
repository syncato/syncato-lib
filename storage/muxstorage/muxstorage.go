package muxstorage

import (
	"errors"
	"fmt"
	"github.com/syncato/syncato-lib/auth"
	"github.com/syncato/syncato-lib/logger"
	"github.com/syncato/syncato-lib/storage"
	"io"
	"net/url"
)

// MuxStorage is a multiplexer for different storages.
// It registers diferent storages and then it routers the operations to the
// corresponsing storage according to the scheme in the path provided in the operation.
// MuxStorage MUST bridge ALL the operations specified in the Storage interface.
type MuxStorage struct {
	registeredStorages map[string]storage.Storage
	log                *logger.Logger
}

// NewMuxStorage receives and array of storages to register.
// It returns a MuxStorage or any error found.
func NewMuxStorage(log *logger.Logger) (*MuxStorage, error) {
	m := MuxStorage{}
	m.registeredStorages = make(map[string]storage.Storage)
	m.log = log
	return &m, nil
}

func (mux *MuxStorage) RegisterStorage(s storage.Storage) error {
	if _, ok := mux.registeredStorages[s.GetScheme()]; ok {
		return errors.New(fmt.Sprintf("storage %s already registered", s.GetScheme()))
	}
	mux.registeredStorages[s.GetScheme()] = s
	return nil
}

// PutFile routes the creation of a file to the corresponding storage.
// It returns any error found.
func (mux *MuxStorage) PutFile(authRes *auth.AuthResource, rawUri string, r io.Reader, size int64) error {
	s, uri, err := mux.getStorageAndURIFromPath(rawUri)
	if err != nil {
		return err
	}
	return s.PutFile(authRes, uri, r, size)
}

func (mux *MuxStorage) GetFile(authRes *auth.AuthResource, rawUri string) (io.Reader, error) {
	s, uri, err := mux.getStorageAndURIFromPath(rawUri)
	if err != nil {
		return nil, err
	}
	return s.GetFile(authRes, uri)
}

func (mux *MuxStorage) Stat(authRes *auth.AuthResource, rawUri string, children bool) (*storage.MetaData, error) {
	s, uri, err := mux.getStorageAndURIFromPath(rawUri)
	if err != nil {
		return nil, err
	}
	return s.Stat(authRes, uri, children)
}

func (mux *MuxStorage) Remove(authRes *auth.AuthResource, rawUri string, recursive bool) error {
	s, uri, err := mux.getStorageAndURIFromPath(rawUri)
	if err != nil {
		return err
	}
	return s.Remove(authRes, uri, recursive)
}

func (mux *MuxStorage) CreateCol(authRes *auth.AuthResource, rawUri string, recursive bool) error {
	s, uri, err := mux.getStorageAndURIFromPath(rawUri)
	if err != nil {
		return err
	}
	return s.CreateCol(authRes, uri, recursive)
}

func (mux *MuxStorage) Copy(authRes *auth.AuthResource, fromRawUri, toRawUri string) error {
	fromStorage, fromUri, err := mux.getStorageAndURIFromPath(fromRawUri)
	if err != nil {
		return err
	}

	toStorage, toUri, err := mux.getStorageAndURIFromPath(toRawUri)
	if err != nil {
		return err
	}

	if fromStorage.GetScheme() != toStorage.GetScheme() {
		return &storage.CrossStorageCopyNotImplemented{}
	}

	// we could use toStorage too, are the same in this step
	return fromStorage.Copy(authRes, fromUri, toUri)
}

func (mux *MuxStorage) Rename(authRes *auth.AuthResource, fromRawUri, toRawUri string) error {
	fromStorage, fromUri, err := mux.getStorageAndURIFromPath(fromRawUri)
	if err != nil {
		return err
	}

	toStorage, toUri, err := mux.getStorageAndURIFromPath(toRawUri)
	if err != nil {
		return err
	}

	if fromStorage.GetScheme() != toStorage.GetScheme() {
		return &storage.CrossStorageMoveNotImplemented{}
	}

	// we could use toStorage too, are the same in this step
	return fromStorage.Rename(authRes, fromUri, toUri)
}

// getStorageFromPath returns the storage associated with the path or any error found.
func (mux *MuxStorage) getStorageAndURIFromPath(path string) (storage.Storage, *url.URL, error) {
	uri, err := url.Parse(path)
	if err != nil {
		return nil, nil, &storage.NotExistError{err.Error()}
	}
	s, ok := mux.registeredStorages[uri.Scheme]
	if !ok {
		return nil, nil, &storage.NotExistError{"no storage registered"}
	}
	return s, uri, nil
}
