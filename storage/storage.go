// Copyright 2015 The Syncato Authors.  All rights reserved.
// Use of this source code is governed by a AGPL
// license that can be found in the LICENSE file.

// Package storage defines the interface that storage providers must implement and
// defines the metadata resource.
package storage

import (
	"github.com/syncato/lib/auth"
	"io"
	"net/url"
)

// StorageProvider is the interface that all the storage providers must implement
// to be used by the storage multiplexer.
// An storage provider is defined by an id called Scheme.
//
// A resource is uniquely identified by a URI http://en.wikipedia.org/wiki/Uniform_resource_identifier
type StorageProvider interface {
	// GetScheme returns the scheme/id of this storage.
	GetScheme() string

	// CreateUserHome creates the user home directory in the storage.
	CreateUserHome(authRes *auth.AuthResource) error

	// IsUserHomeCreated checks if the user home directory has been created or not.
	IsUserHomeCreated(authRes *auth.AuthResource) (bool, error)

	// PutFile puts a file into the storage defined by the uri.
	PutFile(authRes *auth.AuthResource, uri *url.URL, r io.Reader, size int64) error

	// GetFile gets a file from the storage defined by the uri.
	GetFile(authRes *auth.AuthResource, uri *url.URL) (io.Reader, error)

	// Stat returns metadata information about the resources and its children.
	Stat(authRes *auth.AuthResource, uri *url.URL, children bool) (*MetaData, error)

	// Remove removes a resource from the storage defined by the uri.
	Remove(authRes *auth.AuthResource, uri *url.URL, recursive bool) error

	// CreateCol creates a collection in the storage defined by the uri.
	CreateCol(authRes *auth.AuthResource, uri *url.URL, recursive bool) error

	// Copy copies a resource from one uri to another.
	// If uris belong to different storages this is a cross-storage copy.
	Copy(authRes *auth.AuthResource, fromUri, toUri *url.URL) error

	// Rename renames/move a resource from one uri to another.
	// If uris belong to different storages this is a cross-storage rename.
	Rename(authRes *auth.AuthResource, fromUri, toUri *url.URL) error

	// ConvertError convert a storage provider implementation error to the ones defined in this package.
	// This is needed to provide the same logic independently of the storage provider implementation.
	//
	// For example, using a local filesystem the return code "no such file or directory" will be converted
	// to a NotExistError, but if using Swift or Amazon S3, the HTTP 404 error will be converted.
	ConvertError(err error) error

	// GetCapabilities returns the capabilities of this storage.
	GetCapabilities() *Capabilities

	/*
		// SHARE OPERATIONS
		ShareCol(authRes *auth.AuthResource, uri *url.URL, username string, perm *Permissions) error
		ShareByLink(authRes *auth.AuthResource, uri *url.URL, perm *Permissions) error

		// MISC

			Install(v interface{}) error
			GetFile(path string) (io.Reader, error)
			PutFile(path string, r io.Reader, size int64, checksumType string, checksum string) error
			Stat(path string, children bool) (*MetaData, error)
			Remove(path string, recursive bool) error
			CreateCol(path string, recursive bool) error
			Copy(from, to string) error
			Rename(from, to string) error
			GetVersion(path string) (io.Reader, error)
			ListVersions(path string) ([]MetaData, error)
			RollbackVersion(path string) bool

			ListJunkFiles() ([]MetaData, error)
			RestoreJunkFiles(fileIDs []string) error
			PurgeJunkFile(fileID []string) error
			SetupHomeStorage(authRes *auth.AuthResource) error
	*/

}

// MetaData represents the metadata information about a resource.
type MetaData struct {
	Id           string      `json:"id"`            // The id of this resource.
	Path         string      `json:"path"`          // The path of this resource.
	Size         uint64      `json:"size"`          // The size of this resource.
	IsCol        bool        `json:"iscol"`         // Indicates if the resource is a collection.
	MimeType     string      `json:"mime_type"`     // The mimetype of the resource.
	Checksum     string      `json:"checksum"`      // The checksum of the resource.
	ChecksumType string      `json:"checksum_type"` // The type of checksum used to calculate the checksum.
	Modified     uint64      `json:"modified"`      // The latest time the resource has been modified.
	ETag         string      `json:"etag"`          // The ETag http://en.wikipedia.org/wiki/HTTP_ETag.
	Children     []*MetaData `json:"children"`      // If this resource is a collection contains all the children´s metadata.
	Extra        interface{} `json:"extra"`         // Contains extra attributes defined by the storage provider implementation.
}

// Capabilites reprents the capabilities of a storage
// TODO: cross copy-move, versions, ....
type Capabilities struct {
}

type Permissions struct {
	Read   bool
	Write  bool
	Delete bool
}

type ExistError struct {
	Err string
}

func (e *ExistError) Error() string { return e.Err }

type NotExistError struct {
	Err string
}

func (e *NotExistError) Error() string { return e.Err }

type CrossStorageCopyNotImplemented struct {
}

func (e *CrossStorageCopyNotImplemented) Error() string { return "cross storage copy not implemented" }

type CrossStorageMoveNotImplemented struct {
}

func (e *CrossStorageMoveNotImplemented) Error() string { return "cross storage move not implemented" }

func IsExistError(err error) bool {
	_, ok := err.(*ExistError)
	if ok {
		return true
	}
	return false
}

func IsNotExistError(err error) bool {
	_, ok := err.(*NotExistError)
	if ok {

		return true
	}
	return false
}
