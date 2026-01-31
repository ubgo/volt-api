package filesystem

import "os"

const (
	LIST_SHALLOW = false
	LIST_DEEP    = true
)

type DirectoryListing interface{}
type StorageAttributes interface{}

type FilesystemReader interface {

	// fileExists checks if the file exists at the specified location.
	FileExists(location string) (bool, error)

	// directoryExists checks if the directory exists at the specified location.
	DirectoryExists(location string) (bool, error)

	// has checks if the file or directory exists at the specified location.
	Has(location string) (bool, error)

	// read reads the content of a file at the specified location.
	Read(location string) (string, error)

	// readStream reads the content of a file at the specified location as a stream.
	ReadStream(location string) (*os.File, error) // `interface{}` here represents resource

	// listContents lists the contents of the directory at the specified location. `deep` determines the depth of listing.
	ListContents(location string, deep bool) (interface{}, error)

	// lastModified gets the last modified timestamp of the file at the specified path.
	LastModified(path string) (int64, error)

	// fileSize gets the size of the file at the specified path.
	FileSize(path string) (int64, error)

	// mimeType gets the MIME type of the file at the specified path.
	MimeType(path string) (string, error)

	// visibility gets the visibility (permissions) of the file at the specified path.
	Visibility(path string) (string, error)
}

type FilesystemWriter interface {

	// write writes the content to a file at the specified location.
	Write(location string, contents []byte, config map[string]interface{}) error

	// writeStream writes the content to a file at the specified location as a stream.
	WriteStream(location string, contents interface{}, config map[string]interface{}) error

	// setVisibility sets the visibility (permissions) of the file at the specified path.
	SetVisibility(path string, visibility string) error

	// delete deletes the file at the specified location.
	Delete(location string) error

	// deleteDirectory deletes the directory at the specified location.
	DeleteDirectory(location string) error

	// createDirectory creates a directory at the specified location.
	CreateDirectory(location string, config map[string]interface{}) error

	// move moves a file from source to destination.
	Move(source string, destination string, config map[string]interface{}) error

	// copy copies a file from source to destination.
	Copy(source string, destination string, config map[string]interface{}) error
}

type FilesystemOperator interface {
	FilesystemReader
	FilesystemWriter
}

type FilesystemAdapter interface {
	// fileExists checks if the file exists at the specified location.
	FileExists(path string) (bool, error)

	// directoryExists checks if the directory exists at the specified location.
	DirectoryExists(path string) (bool, error)

	// write writes the content to a file at the specified location.
	Write(path string, contents []byte, config map[string]interface{}) error

	// writeStream writes the content to a file at the specified location as a stream.
	WriteStream(path string, contents interface{}, config map[string]interface{}) error

	// read reads the content of a file at the specified location.
	Read(path string) (string, error)

	// readStream reads the content of a file at the specified location as a stream.
	ReadStream(path string) (*os.File, error) // `interface{}` here represents resource

	// delete deletes the file at the specified location.
	Delete(path string) error

	// deleteDirectory deletes the directory at the specified location.
	DeleteDirectory(path string) error

	// createDirectory creates a directory at the specified location.
	CreateDirectory(path string, config map[string]interface{}) error

	// setVisibility sets the visibility (permissions) of the file at the specified path.
	SetVisibility(path string, visibility string) error

	// visibility gets the visibility (permissions) of the file at the specified path.
	Visibility(path string) (string, error)

	// mimeType gets the MIME type of the file at the specified path.
	MimeType(path string) (string, error)

	// lastModified gets the last modified timestamp of the file at the specified path.
	LastModified(path string) (int64, error)

	// fileSize gets the size of the file at the specified path.
	FileSize(path string) (int64, error)

	// listContents lists the contents of the directory at the specified location. `deep` determines the depth of listing.
	ListContents(path string, deep bool) (interface{}, error)

	// move moves a file from source to destination.
	Move(source string, destination string, config map[string]interface{}) error

	// copy copies a file from source to destination.
	Copy(source string, destination string, config map[string]interface{}) error
}

// PathNormalizer is an interface for normalizing paths.
type PathNormalizer interface {
	// NormalizePath normalizes the given path.
	NormalizePath(path string) (string, error)
}
