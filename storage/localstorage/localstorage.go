package localstorage

import (
	"fmt"
	"github.com/syncato/syncato-lib/auth"
	"github.com/syncato/syncato-lib/config"
	"github.com/syncato/syncato-lib/logger"
	"github.com/syncato/syncato-lib/storage"
	"io"
	"mime"
	"net/url"
	"os"
	"path/filepath"
)

type LocalStorage struct {
	scheme      string
	cp          *config.ConfigProvider
	log         *logger.Logger
	rootDataDir string
	rootTmpDir  string
}

func NewLocalStorage(scheme string, cp *config.ConfigProvider, log *logger.Logger) (*LocalStorage, error) {
	s := &LocalStorage{scheme: scheme, cp: cp, log: log}
	cfg, err := cp.ParseFile()
	if err != nil {
		return nil, err
	}
	s.rootDataDir = cfg.RootDataDir
	s.rootTmpDir = cfg.RootTmpDir
	return s, nil
}

func (s *LocalStorage) GetScheme() string {
	return s.scheme
}

func (s *LocalStorage) PutFile(authRes *auth.AuthResource, uri *url.URL, r io.Reader, size int64) error {
	tmpPath := filepath.Join(s.rootTmpDir, filepath.Base(uri.Path))

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

func (s *LocalStorage) Stat(authRes *auth.AuthResource, uri *url.URL, children bool) (*storage.MetaData, error) {
	absPath := filepath.Clean(filepath.Join(s.rootDataDir, uri.Path))

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
		Size:     finfo.Size(),
		IsCol:    finfo.IsDir(),
		Modified: finfo.ModTime().Unix(),
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
			Size:     f.Size(),
			IsCol:    f.IsDir(),
			Modified: f.ModTime().Unix(),
			ETag:     fmt.Sprintf("\"%d\"", f.ModTime().Unix()),
			MimeType: mimeType,
		}
		meta.Children[i] = &m
	}

	return &meta, nil
}

func (s *LocalStorage) GetFile(authRes *auth.AuthResource, uri *url.URL) (io.Reader, error) {
	absPath := filepath.Clean(filepath.Join(s.rootDataDir, uri.Path))
	file, err := os.Open(absPath)
	if err != nil {
		return nil, s.ConvertError(err)
	}
	return file, nil
}

func (s *LocalStorage) Remove(authRes *auth.AuthResource, uri *url.URL, recursive bool) error {
	absPath := filepath.Clean(filepath.Join(s.rootDataDir, uri.Path))
	if !recursive {
		return s.ConvertError(os.Remove(absPath))
	}
	return s.ConvertError(os.RemoveAll(absPath))
}

func (s *LocalStorage) CreateCol(authRes *auth.AuthResource, uri *url.URL, recursive bool) error {
	absPath := filepath.Clean(filepath.Join(s.rootDataDir, uri.Path))
	if recursive == false {
		return s.ConvertError(os.Mkdir(absPath, 0666))
	}
	return s.ConvertError(os.MkdirAll(absPath, 0666))
}

func (s *LocalStorage) Copy(authRes *auth.AuthResource, fromUri, toUri *url.URL) error {
	fromabsPath := filepath.Clean(filepath.Join(s.rootDataDir, fromUri.Path))
	toabsPath := filepath.Clean(filepath.Join(s.rootDataDir, toUri.Path))
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

func (s *LocalStorage) Rename(authRes *auth.AuthResource, fromUri, toUri *url.URL) error {
	fromabsPath := filepath.Clean(filepath.Join(s.rootDataDir, fromUri.Path))
	toabsPath := filepath.Clean(filepath.Join(s.rootDataDir, toUri.Path))
	return s.ConvertError(os.Rename(fromabsPath, toabsPath))
}

func (s *LocalStorage) ConvertError(err error) error {
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

func (s *LocalStorage) GetCapabilities() *storage.Capabilities {
	cap := storage.Capabilities{}
	return &cap
}

func (s *LocalStorage) commitPutFile(from, to string) error {
	toabsPath := filepath.Join(s.rootDataDir, to)
	return os.Rename(from, toabsPath)
}
