package filesystem

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

var fileSystemAdapter = &LocalFileSystemAdapter{}

func Get() *LocalFileSystemAdapter {
	return fileSystemAdapter
}

type LocalFileSystemAdapter struct {
	rootPath        string
	rootPathInSetup bool
	prefixer        PathPrefixer
}

func NewLocalFileSystemAdapter(rootPath string) *LocalFileSystemAdapter {

	prefixer := NewPathPrefixer(rootPath, "/")
	lfs := &LocalFileSystemAdapter{rootPath: rootPath, prefixer: *prefixer}

	// lfs.EnsureRootDirectoryExists()

	return lfs
}

func (lfs *LocalFileSystemAdapter) EnsureRootDirectoryExists() {
	panic("not implemented")
}

func (lfs *LocalFileSystemAdapter) EnsuretDirectoryExists(dirname string, visibility int) error {
	panic("not implemented")

	if lfs.rootPathInSetup {
		return nil
	}

	lfs.rootPathInSetup = true

	return nil
}

func (lfs *LocalFileSystemAdapter) FileExists(location string) (bool, error) {
	info, err := os.Stat(location)

	if err == nil {
		return !info.IsDir(), nil
	}

	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func (lfs *LocalFileSystemAdapter) DirectoryExists(location string) (bool, error) {
	info, err := os.Stat(location)
	if err == nil && info.IsDir() {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// func (lfs *LocalFileSystemAdapter) Has(location string) (bool, error) {
// 	return lfs.FileExists(location)
// }

func (lfs *LocalFileSystemAdapter) Read(location string) (string, error) {
	content, err := os.ReadFile(location)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func (lfs *LocalFileSystemAdapter) ReadStream(location string) (*os.File, error) {
	file, err := os.Open(location)
	if err != nil {
		return nil, err
	}
	return file, nil
}

func (lfs *LocalFileSystemAdapter) ListContents(location string, deep bool) (interface{}, error) {
	var fileList []os.FileInfo

	err := filepath.Walk(location, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if deep || filepath.Dir(path) == location {
			fileList = append(fileList, info)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return fileList, nil
}

func (lfs *LocalFileSystemAdapter) LastModified(path string) (int64, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return info.ModTime().Unix(), nil
}

func (lfs *LocalFileSystemAdapter) FileSize(path string) (int64, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

// YTD
func (lfs *LocalFileSystemAdapter) MimeType(path string) (string, error) {
	// To be implemented based on MIME type detection library/method used
	return "", nil
}

func (lfs *LocalFileSystemAdapter) Visibility(path string) (string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	mode := info.Mode().Perm()
	return mode.String(), nil
}

func (lfs *LocalFileSystemAdapter) Write(path string, contents []byte, config map[string]interface{}) error {

	// YTD - ensure root Directory exists

	// YTD - ensure direcotry exists

	prefixedPath := lfs.prefixer.PrefixPath(path)

	file, err := os.Create(prefixedPath)
	if err != nil {
		return err
	}

	_, err = file.Write(contents)

	file.Close()
	// YTD - set visibility
	return err
}

// YTD
func (lfs *LocalFileSystemAdapter) WriteStream(location string, contents interface{}, config map[string]interface{}) error {
	file, err := os.OpenFile(location, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	switch v := contents.(type) {
	case []byte:
		_, err = file.Write(v)
	case string:
		_, err = file.WriteString(v)
	default:
		return fmt.Errorf("unsupported content type")
	}
	return err
}

// YTD
func (lfs *LocalFileSystemAdapter) SetVisibility(path string, visibility string) error {
	var mode os.FileMode
	switch visibility {
	case "public":
		mode = 0644 // -rw-r--r--
	case "private":
		mode = 0600 // -rw-------
	default:
		return fmt.Errorf("unsupported visibility type")
	}

	return os.Chmod(path, mode)
}

func (lfs *LocalFileSystemAdapter) Delete(path string) error {
	prefixedPath := lfs.prefixer.PrefixPath(path)

	if _, err := os.Stat(prefixedPath); errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return os.Remove(prefixedPath)
}

// YTD
func (lfs *LocalFileSystemAdapter) DeleteDirectory(location string) error {
	return os.RemoveAll(location)
}

// YTD
func (lfs *LocalFileSystemAdapter) CreateDirectory(location string, config map[string]interface{}) error {
	return os.MkdirAll(location, 0755) // rwxr-xr-x
}

// YTD
func (lfs *LocalFileSystemAdapter) Move(source string, destination string, config map[string]interface{}) error {
	return os.Rename(source, destination)
}

// YTD
func (lfs *LocalFileSystemAdapter) Copy(source string, destination string, config map[string]interface{}) error {
	sourceFile, err := os.Open(source)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destinationFile, err := os.Create(destination)
	if err != nil {
		return err
	}
	defer destinationFile.Close()

	_, err = io.Copy(destinationFile, sourceFile)
	return err
}
