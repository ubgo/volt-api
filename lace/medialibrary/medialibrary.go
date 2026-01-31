package medialibrary

import (
	"bytes"
	"context"
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"fmt"
	"lace/filesystem"
	"lace/publicid"
	"log"
	"mime"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
)

var db *bun.DB

const (
	VARIANT_NAME_ORIGINAL = "original"
)

type Media struct {
	ID            string
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
	Disk          string    `json:"disk"`
	Directory     string    `json:"directory"`
	Name          string    `json:"name"`
	OriginalName  string    `json:"originalName"`
	Extension     string    `json:"extension"`
	MimeType      string    `json:"mimeType"`
	AggregateType string    `json:"aggregateType"`
	Uid           string    `json:"uid"`
	Checksum      string    `json:"checksum"`
	IsVariant     bool      `json:"isVariant"`
	WorkspaceID   string    `json:"workspaceId"`
	Size          uint      `json:"size"`
	URL           string    `json:"url" bun:"-"`
}

func (m *Media) FilePath() string {
	if len(m.Directory) == 0 {
		return m.Name
	}
	return m.Directory + m.Name
}

func Construct(sqldb *sql.DB) {
	InitDb(sqldb)
}

func InitDb(sqldb *sql.DB) {
	db = bun.NewDB(sqldb, pgdialect.New())
}

// Responsible only for single Media
type FileAdder struct {
	rootPath    string
	filename    string
	contentType string
	filesystem  *filesystem.FileSystem
}

func NewFileAdder(rootPath string) *FileAdder {
	if db == nil {
		panic("medialibrary not constructed yet")
	}

	adpater := filesystem.NewLocalFileSystemAdapter(rootPath)
	fs := filesystem.NewFileSystem(adpater, filesystem.WithPublicURL([]string{"http://localhost/uploads"}))

	return &FileAdder{
		filesystem: fs,
		rootPath:   rootPath,
	}
}

// GetChecksum computes and returns the MD5 checksum of the provided byte slice.
func (cls *FileAdder) GetChecksum(b []byte) string {
	checkSum := md5.Sum(b)
	return hex.EncodeToString(checkSum[:])
}

func (cls *FileAdder) CreateFromMultiPart(filePart *multipart.Part) (*Media, error) {
	fmt.Println(filePart.Header)
	_, params, err := mime.ParseMediaType(filePart.Header.Get("Content-Disposition"))
	if err != nil {
		log.Printf("Invalid Content-Disposition: %v", err)
		return nil, err
	}

	cls.filename = params["filename"]
	cls.contentType = filePart.Header.Get("Content-Type")

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(filePart)
	if err != nil {
		return nil, err
	}

	b := buf.Bytes()

	err = cls.filesystem.Write(cls.filename, b, nil)

	if err != nil {
		return nil, err
	}

	ext := strings.ToLower(filepath.Ext(cls.filename))

	media := Media{
		ID:           publicid.Must12(),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		Checksum:     cls.GetChecksum(b),
		Name:         cls.filename, // actual filename of the file saved
		OriginalName: cls.filename, // is the name of file uploaded by user but we can save it filename-xxx.jpeg (name)
		MimeType:     cls.contentType,
		Disk:         "local",
		// Directory:    cls.rootPath,
		Directory:   "",
		IsVariant:   false,
		Uid:         publicid.Must14(),
		Extension:   ext[1:],
		Size:        uint(len(b)), // in bytes
		WorkspaceID: "default",
	}

	// goutil.PrintToJSON(media)

	_, err = db.NewInsert().Model(&media).Exec(context.Background())

	if err != nil {
		cls.filesystem.Delete(cls.filename)
		return nil, err
	}

	media.URL, _ = cls.filesystem.PublicURL(media.FilePath())

	// fmt.Println("res", res)
	return &media, nil
}
