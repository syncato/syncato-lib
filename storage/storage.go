package storage

import (
	"github.com/syncato/syncato-lib/auth"
	"io"
	"net/url"
)

// Storage is the interface that wraps the basic storage methods.
type Storage interface {
	GetScheme() string

	// FS OPERATIONS
	PutFile(authRes *auth.AuthResource, uri *url.URL, r io.Reader, size int64) error
	GetFile(authRes *auth.AuthResource, uri *url.URL) (io.Reader, error)
	Stat(authRes *auth.AuthResource, uri *url.URL, children bool) (*MetaData, error)
	Remove(authRes *auth.AuthResource, uri *url.URL, recursive bool) error
	CreateCol(authRes *auth.AuthResource, uri *url.URL, recursive bool) error
	Copy(authRes *auth.AuthResource, fromUri, toUri *url.URL) error
	Rename(authRes *auth.AuthResource, fromUri, toUri *url.URL) error

	ConvertError(err error) error
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

// A MetaData represents the metadata information about a resource.
type MetaData struct {
	Id           string      `json:"id"`
	Path         string      `json:"path"`
	Size         int64       `json:"size"`
	IsCol        bool        `json:"isCol"`
	MimeType     string      `json:"mimeType"`
	Checksum     string      `json:"checksum"`
	ChecksumType string      `json:"checksumType"`
	Modified     int64       `json:"modified"`
	ETag         string      `json:"etag"`
	Children     []*MetaData `json:"children"`
	Extra        interface{} `json:"extra"` // maybe to save xattrs or custom user data
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
