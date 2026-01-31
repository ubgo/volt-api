package filesystem

import "fmt"

type FileSystemManager struct {
	disk map[string]*FileSystem
}

func NewFileSystemManager() *FileSystemManager {
	return &FileSystemManager{disk: make(map[string]*FileSystem)}
}

func (m *FileSystemManager) AddDisk(name string, fs *FileSystem) {
	m.disk[name] = fs
}

func (m *FileSystemManager) GetDisk(name string) (*FileSystem, error) {
	if m.disk[name] == nil {
		return nil, fmt.Errorf("no disk found: %s", name)
	}
	return m.disk[name], nil
}
