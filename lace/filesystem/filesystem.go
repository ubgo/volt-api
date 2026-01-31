package filesystem

import (
	"fmt"
	"os"
)

type FileSystem struct {
	adapter        FilesystemAdapter
	pathNormalizer PathNormalizer
	cfg            configFileSystem
}

func NewFileSystem(adapter FilesystemAdapter, opts ...FileSystemOption) *FileSystem {
	cfg := configFileSystem{}
	cfg.options(opts...)

	return &FileSystem{adapter: adapter, pathNormalizer: &WhitespacePathNormalizer{}, cfg: cfg}
}

func (cls *FileSystem) FileExists(location string) (bool, error) {
	npath, err := cls.pathNormalizer.NormalizePath(location)
	if err != nil {
		return false, err
	}
	return cls.adapter.FileExists(npath)
}

func (cls *FileSystem) DirectoryExists(location string) (bool, error) {
	npath, err := cls.pathNormalizer.NormalizePath(location)
	if err != nil {
		return false, err
	}
	return cls.adapter.DirectoryExists(npath)
}

func (cls *FileSystem) Has(location string) (bool, error) {
	fileExists, err := cls.FileExists(location)
	if err != nil {
		return false, err
	}
	dirExists, errDir := cls.DirectoryExists(location)
	if errDir != nil {
		return false, errDir
	}

	return fileExists || dirExists, nil
}

func (cls *FileSystem) Read(location string) (string, error) {
	npath, err := cls.pathNormalizer.NormalizePath(location)
	if err != nil {
		return "", err
	}
	return cls.adapter.Read(npath)
}

func (cls *FileSystem) ReadStream(location string) (*os.File, error) {
	npath, err := cls.pathNormalizer.NormalizePath(location)
	if err != nil {
		return nil, err
	}
	return cls.adapter.ReadStream(npath)
}

// YTD
func (cls *FileSystem) ListContents(location string, deep bool) (interface{}, error) {
	panic("not implemented")
}

func (cls *FileSystem) LastModified(path string) (int64, error) {
	npath, err := cls.pathNormalizer.NormalizePath(path)
	if err != nil {
		return 0, err
	}
	return cls.adapter.LastModified(npath)
}

func (cls *FileSystem) FileSize(path string) (int64, error) {
	npath, err := cls.pathNormalizer.NormalizePath(path)
	if err != nil {
		return 0, err
	}
	return cls.adapter.FileSize(npath)
}

func (cls *FileSystem) MimeType(path string) (string, error) {
	npath, err := cls.pathNormalizer.NormalizePath(path)
	if err != nil {
		return "", err
	}
	return cls.adapter.MimeType(npath)
}

func (cls *FileSystem) Visibility(path string) (string, error) {
	npath, err := cls.pathNormalizer.NormalizePath(path)
	if err != nil {
		return "", err
	}
	return cls.adapter.Visibility(npath)
}

func (cls *FileSystem) Write(location string, contents []byte, config map[string]interface{}) error {
	npath, err := cls.pathNormalizer.NormalizePath(location)
	if err != nil {
		return err
	}
	return cls.adapter.Write(npath, contents, config)
}

// YTD
func (cls *FileSystem) WriteStream(location string, contents interface{}, config map[string]interface{}) error {
	panic("not implemented")
}

func (cls *FileSystem) SetVisibility(path string, visibility string) error {
	npath, err := cls.pathNormalizer.NormalizePath(path)
	if err != nil {
		return err
	}
	return cls.adapter.SetVisibility(npath, visibility)
}

func (cls *FileSystem) Delete(location string) error {
	npath, err := cls.pathNormalizer.NormalizePath(location)
	if err != nil {
		return err
	}
	return cls.adapter.Delete(npath)
}

func (cls *FileSystem) DeleteDirectory(location string) error {
	npath, err := cls.pathNormalizer.NormalizePath(location)
	if err != nil {
		return err
	}
	return cls.adapter.DeleteDirectory(npath)
}

func (cls *FileSystem) CreateDirectory(location string, config map[string]interface{}) error {
	npath, err := cls.pathNormalizer.NormalizePath(location)
	if err != nil {
		return err
	}
	return cls.adapter.CreateDirectory(npath, config)
}

// YTD
func (cls *FileSystem) Move(source string, destination string, config map[string]interface{}) error {
	panic("not implemented")
}

// YTD
func (cls *FileSystem) Copy(source string, destination string, config map[string]interface{}) error {
	panic("not implemented")
}

func (cls *FileSystem) PublicURL(path string) (string, error) {

	if cls.cfg.PublicURLGenerator != nil {
		return cls.cfg.PublicURLGenerator(path)
	}

	if len(cls.cfg.PublicURL) > 0 {
		pathPrefixer := NewPrefixPublicUrlGenerator(cls.cfg.PublicURL[0])
		return pathPrefixer.PublicURL(path)

	}

	return "", fmt.Errorf("not implemented")
}
