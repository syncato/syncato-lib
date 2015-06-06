// Copyright 2015 The Syncato Authors.  All rights reserved.
// Use of this source code is governed by a AGPL
// license that can be found in the LICENSE file.

// Package mux defines the storage multiplexer to route storage operations against
// the registered storage providers.
package mux

import (
	"errors"
	"fmt"
	"github.com/syncato/lib/auth"
	"github.com/syncato/lib/logger"
	"github.com/syncato/lib/storage"
	"io"
	"net/url"
)

// StorageMux is a multiplexer responsible for routing storage operations to the
// correct storage provider implementation.
// It keeps a map of all the storage providers registered.
type StorageMux struct {
	storageProviders map[string]storage.StorageProvider
	log              *logger.Logger
}

// NewStorageMux creates a StorageMux or returns an error
func NewStorageMux(log *logger.Logger) (*StorageMux, error) {
	m := StorageMux{}
	m.storageProviders = make(map[string]storage.StorageProvider)
	m.log = log
	return &m, nil
}

// AddStorageProvider adds a storage provider to be used by the multiplexer.
func (mux *StorageMux) AddStorageProvider(s storage.StorageProvider) error {
	if _, ok := mux.storageProviders[s.GetScheme()]; ok {
		return errors.New(fmt.Sprintf("storage %s already registered", s.GetScheme()))
	}
	mux.storageProviders[s.GetScheme()] = s
	return nil
}

func (mux *StorageMux) GetStorageProvider(storageScheme string) (storage.StorageProvider, bool) {
	sp, ok := mux.storageProviders[storageScheme]
	return sp, ok
}

// IsUserHomeCreated checks if the user home directory has been created in the specified storage.
func (mux *StorageMux) IsUserHomeCreated(authRes *auth.AuthResource, storageScheme string) (bool, error) {
	storage, ok := mux.GetStorageProvider(storageScheme)
	if !ok {
		return false, errors.New(fmt.Sprintf("storage '%s' not registered", storageScheme))
	}
	ok, err := storage.IsUserHomeCreated(authRes)
	if err != nil {
		return false, err
	}
	if !ok {
		return false, nil
	}
	return true, nil
}

// CreateUserHome routes the creation of the user home directory to the correct storage provider implementation.
// If the storageScheme is empty, the creation of the home directory will be propagated to all storages.
func (mux *StorageMux) CreateUserHome(authRes *auth.AuthResource, storageScheme string) error {
	storage, ok := mux.GetStorageProvider(storageScheme)
	if !ok {
		return errors.New(fmt.Sprintf("storage '%s' not registered", storageScheme))
	}
	return storage.CreateUserHome(authRes)
}

// PutFile routes the put operation to the correct storage provider implementation.
func (mux *StorageMux) PutFile(authRes *auth.AuthResource, rawUri string, r io.Reader, size int64) error {
	s, uri, err := mux.getStorageAndURIFromPath(rawUri)
	if err != nil {
		return err
	}
	return s.PutFile(authRes, uri, r, size)
}

// GetFile routes the get operation to the correct storage provider implementation.
func (mux *StorageMux) GetFile(authRes *auth.AuthResource, rawUri string) (io.Reader, error) {
	s, uri, err := mux.getStorageAndURIFromPath(rawUri)
	if err != nil {
		return nil, err
	}
	return s.GetFile(authRes, uri)
}

// Stat routes the stat operation to the correct storage provider implementation.
func (mux *StorageMux) Stat(authRes *auth.AuthResource, rawUri string, children bool) (*storage.MetaData, error) {
	mux.log.Debug(fmt.Sprintf("%+v", children), nil)
	s, uri, err := mux.getStorageAndURIFromPath(rawUri)
	if err != nil {
		return nil, err
	}
	return s.Stat(authRes, uri, children)
}

// Remove routes the remove operation to the correct storage provider implementation.
func (mux *StorageMux) Remove(authRes *auth.AuthResource, rawUri string, recursive bool) error {
	s, uri, err := mux.getStorageAndURIFromPath(rawUri)
	if err != nil {
		return err
	}
	return s.Remove(authRes, uri, recursive)
}

// CreateCol routes the create collection operation to the correct storage provider implementation.
func (mux *StorageMux) CreateCol(authRes *auth.AuthResource, rawUri string, recursive bool) error {
	s, uri, err := mux.getStorageAndURIFromPath(rawUri)
	if err != nil {
		return err
	}
	return s.CreateCol(authRes, uri, recursive)
}

// Copy routes the copy operation to the correct storage provider implementation.
func (mux *StorageMux) Copy(authRes *auth.AuthResource, fromRawUri, toRawUri string) error {
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

	return fromStorage.Copy(authRes, fromUri, toUri)
}

// Rename routes the rename operation to the correct storage provider implementation.
func (mux *StorageMux) Rename(authRes *auth.AuthResource, fromRawUri, toRawUri string) error {
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

// getStorageFromPath returns the storage provider adn the URI associated with the resourceUrl passsed or an error.
// the resourceUrl must be a well-formed URI like local://photos/beach.png or eos://data/big.dat
func (mux *StorageMux) getStorageAndURIFromPath(resourceUrl string) (storage.StorageProvider, *url.URL, error) {
	uri, err := url.Parse(resourceUrl)
	if err != nil {
		return nil, nil, &storage.NotExistError{err.Error()}
	}
	s, ok := mux.GetStorageProvider(uri.Scheme)
	if !ok {
		return nil, nil, &storage.NotExistError{fmt.Sprintf("storage %s not registered", uri.Scheme)}
	}
	mux.log.Debug("get storage and uri from url", map[string]interface{}{"url": resourceUrl, "uri": fmt.Sprintf("%+v", *uri)})
	return s, uri, nil
}
