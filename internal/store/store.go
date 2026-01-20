package store

import (
	"fmt"
	"os"
	"path/filepath"
)

type Store struct {
	dataDir string
}

func New(dataDir string) (*Store, error) {
	s := &Store{dataDir: dataDir}
	if err := s.init(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Store) init() error {
	dirs := []string{
		filepath.Join(s.dataDir, "daily"),
		filepath.Join(s.dataDir, "meta"),
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}
	return nil
}

func (s *Store) DataDir() string {
	return s.dataDir
}

func (s *Store) DailyDir() string {
	return filepath.Join(s.dataDir, "daily")
}

func (s *Store) MetaDir() string {
	return filepath.Join(s.dataDir, "meta")
}
