// Copyright 2015 The Syncato Authors.  All rights reserved.
// Use of this source code is governed by a AGPL
// license that can be found in the LICENSE file.

// Package local implements the StorageProvider interface to use a local filesystem as a storage backend.
package local

import (
	"fmt"
	"github.com/syncato/lib/auth"
	"github.com/syncato/lib/config"
	"github.com/syncato/lib/logger"
	"github.com/syncato/lib/storage"
	"io"
	"mime"
	"net/url"
	"os"
	"path/filepath"
)

// StorageLocal is the implementation of the StorageProvider interface to use a local
// filesystem as the storage backend.
type StorageLocal struct {
	scheme      string
	cfg         *config.Config
	log         *logger.Logger
	rootDataDir string
	rootTmpDir  string
}

// NewStorageLocal creates a StorageLocal object or returns an error.
func NewStorageLocal(scheme string, cfg *config.Config, log *logger.Logger) (*StorageLocal, error) {
	s := &StorageLocal{scheme: scheme, cfg: cfg, log: log}
	s.rootDataDir = cfg.RootDataDir()
	s.rootTmpDir = cfg.RootTmpDir()
	return s, nil
}

func (s *StorageLocal) GetScheme() string {
	return s.scheme
}

func (s *StorageLocal) CreateUserHome(authRes *auth.AuthResource) error {
	exists, err := s.IsUserHomeCreated(authRes)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	homeDir := filepath.Join(s.rootDataDir, authRes.AuthID, authRes.Username)
	return os.MkdirAll(homeDir, 0666)
}

func (s *StorageLocal) IsUserHomeCreated(authRes *auth.AuthResource) (bool, error) {
	homeDir := filepath.Join(s.rootDataDir, authRes.AuthID, authRes.Username)
	_, err := os.Stat(homeDir)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func (s *StorageLocal) PutFile(authRes *auth.AuthResource, uri *url.URL, r io.Reader, size int64) error {
	tmpPath := filepath.Join(s.rootTmpDir, authRes.AuthID, authRes.Username, filepath.Base(uri.Path))

	fd, err := os.Create(tmpPath)
	defer fd.Close()
	if err != nil {
		return s.ConvertError(err)
	}
	_, err = io.Copy(fd, r)
	if err != nil {
		return s.ConvertError(err)
	}
	return s.commitPutFile(tmpPath, uri.Path)
}

func (s *StorageLocal) Stat(authRes *auth.AuthResource, uri *url.URL, children bool) (*storage.MetaData, error) {
	absPath := filepath.Clean(filepath.Join(s.rootDataDir, authRes.AuthID, authRes.Username, uri.Path))

	finfo, err := os.Stat(absPath)
	if err != nil {
		return nil, s.ConvertError(err)
	}

	mimeType := mime.TypeByExtension(filepath.Ext(uri.Path))
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}
	if finfo.IsDir() {
		mimeType = "inode/directory"
	}
	meta := storage.MetaData{
		Id:       uri.String(),
		Path:     uri.String(),
		Size:     uint64(finfo.Size()),
		IsCol:    finfo.IsDir(),
		Modified: uint64(finfo.ModTime().Unix()),
		ETag:     fmt.Sprintf("\"%d\"", finfo.ModTime().Unix()),
		MimeType: mimeType,
	}

	if meta.IsCol == false {
		return &meta, nil
	}
	if children == false {
		return &meta, nil
	}

	fd, err := os.Open(absPath)
	defer fd.Close()
	if err != nil {
		return nil, s.ConvertError(err)
	}

	finfos, err := fd.Readdir(0)
	if err != nil {
		return nil, s.ConvertError(err)
	}

	meta.Children = make([]*storage.MetaData, len(finfos))
	for i, f := range finfos {
		uri.Fragment = ""
		uri.RawQuery = ""
		childPath := filepath.Join(uri.String(), f.Name())
		mimeType := mime.TypeByExtension(filepath.Ext(childPath))
		if mimeType == "" {
			mimeType = "application/octet-stream"
		}
		if f.IsDir() {
			mimeType = "inode/directory"
		}
		m := storage.MetaData{
			Id:       childPath,
			Path:     childPath,
			Size:     uint64(f.Size()),
			IsCol:    f.IsDir(),
			Modified: uint64(f.ModTime().Unix()),
			ETag:     fmt.Sprintf("\"%d\"", f.ModTime().Unix()),
			MimeType: mimeType,
		}
		meta.Children[i] = &m
	}

	return &meta, nil
}

func (s *StorageLocal) GetFile(authRes *auth.AuthResource, uri *url.URL) (io.Reader, error) {
	absPath := filepath.Clean(filepath.Join(s.rootDataDir, authRes.AuthID, authRes.Username, uri.Path))
	file, err := os.Open(absPath)
	if err != nil {
		return nil, s.ConvertError(err)
	}
	return file, nil
}

func (s *StorageLocal) Remove(authRes *auth.AuthResource, uri *url.URL, recursive bool) error {
	absPath := filepath.Clean(filepath.Join(s.rootDataDir, authRes.AuthID, authRes.Username, uri.Path))
	if !recursive {
		return s.ConvertError(os.Remove(absPath))
	}
	return s.ConvertError(os.RemoveAll(absPath))
}

func (s *StorageLocal) CreateCol(authRes *auth.AuthResource, uri *url.URL, recursive bool) error {
	absPath := filepath.Clean(filepath.Join(s.rootDataDir, authRes.AuthID, authRes.Username, uri.Path))
	if recursive == false {
		return s.ConvertError(os.Mkdir(absPath, 0666))
	}
	return s.ConvertError(os.MkdirAll(absPath, 0666))
}

func (s *StorageLocal) Copy(authRes *auth.AuthResource, fromUri, toUri *url.URL) error {
	fromabsPath := filepath.Clean(filepath.Join(s.rootDataDir, authRes.AuthID, authRes.Username, fromUri.Path))
	toabsPath := filepath.Clean(filepath.Join(s.rootDataDir, authRes.AuthID, authRes.Username, toUri.Path))
	src, err := os.Open(fromabsPath)
	defer src.Close()
	if err != nil {
		return err
	}
	dst, err := os.Create(toabsPath)
	defer dst.Close()
	if err != nil {
		return err
	}
	_, err = io.Copy(dst, src)
	return err
}

func (s *StorageLocal) Rename(authRes *auth.AuthResource, fromUri, toUri *url.URL) error {
	fromabsPath := filepath.Clean(filepath.Join(s.rootDataDir, authRes.AuthID, authRes.Username, fromUri.Path))
	toabsPath := filepath.Clean(filepath.Join(s.rootDataDir, authRes.AuthID, authRes.Username, toUri.Path))
	return s.ConvertError(os.Rename(fromabsPath, toabsPath))
}

func (s *StorageLocal) ConvertError(err error) error {
	if err == nil {
		return nil
	} else if os.IsExist(err) {
		return &storage.ExistError{err.Error()}
	} else if os.IsNotExist(err) {
		return &storage.NotExistError{err.Error()}
	} else {
		return err
	}
}

func (s *StorageLocal) GetCapabilities() *storage.Capabilities {
	cap := storage.Capabilities{}
	return &cap
}

func (s *StorageLocal) commitPutFile(from, to string) error {
	toabsPath := filepath.Join(s.rootDataDir, to)
	return os.Rename(from, toabsPath)
}
