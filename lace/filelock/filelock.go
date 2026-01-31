package filelock

import (
	"errors"
	"os"
)

type FileLock struct {
	key string
	dir string
}

func New(key string) *FileLock {
	flock := &FileLock{
		key: key,
	}

	return flock
}

func (l *FileLock) SetDir(dir string) *FileLock {
	l.dir = dir
	return l
}

func (l *FileLock) Filename() string {
	return l.key + ".lock"
}

func (l *FileLock) FilePath() string {
	return l.dir + "/" + l.Filename()
}

func (l *FileLock) Lock() error {
	f, err := os.Create(l.FilePath())
	if err != nil {
		return errors.New("cannot create lock file: " + err.Error())
	}
	defer f.Close()

	return nil
}

func (l *FileLock) Release() error {
	if l.IsActive() {
		return os.Remove(l.FilePath())
	}

	return nil
}

func (l *FileLock) IsActive() bool {
	info, err := os.Stat(l.FilePath())
	if err == nil {
		return !info.IsDir()
	}
	if errors.Is(err, os.ErrNotExist) {
		return false
	}
	return false
}
