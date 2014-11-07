package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/nu7hatch/gouuid"
)

var StoreFileError error = errors.New("Store file error")

var filesDir = "store"
var dbName = "db"

type Store interface {
	StoreFile(reader io.Reader) (string, error)
	GetPath() string
}

type fileDescriptor struct {
	Path     string    `json:"path"`
	Uploaded time.Time `json:"uploaded"`
	Size     int64     `json:"size"`
}

type fileStore struct {
	sync.RWMutex
	path  string
	keys  chan string
	items map[string]fileDescriptor
}

func (s *fileStore) String() string {
	return fmt.Sprintf("Store{%s}", s.path)
}

func (s *fileStore) StoreFile(reader io.Reader) (string, error) {
	key := <-s.keys
	filePath := path.Join(s.GetPath(), key)
	file, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()
	written, cpErr := io.Copy(file, reader)
	if cpErr != nil {
		return key, cpErr
	}
	s.Lock()
	s.items[key] = fileDescriptor{Path: filePath, Uploaded: time.Now(), Size: written}
	s.Unlock()
	s.flushDb()
	return key, nil
}

func (s *fileStore) GetPath() string {
	return path.Join(s.path, filesDir)
}

func (s *fileStore) loadDb() {
	f, err := os.Open(path.Join(s.path, dbName))
	if os.IsNotExist(err) {
		return
	}
	if err != nil {
		log.Fatal("Can't open store database: %v", err)
	}
	defer f.Close()
	encoder := json.NewDecoder(f)
	if err := encoder.Decode(&s.items); err != nil {
		log.Fatal("Can't open store database")
	}
}

func (s *fileStore) flushDb() {
	s.RLock()
	defer s.RUnlock()
	f, err := os.Create(path.Join(s.path, dbName))
	if err != nil {
		log.Fatal("Can't open store database")
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	if err := enc.Encode(s.items); err != nil {
		log.Fatal("Can't save store database")
	}
}

func (s *fileStore) init() error {
	storePath := s.GetPath()
	if _, err := os.Stat(storePath); os.IsNotExist(err) {
		if err := os.MkdirAll(storePath, 0755); err != nil {
			log.Fatal("Can't create store directory")
		}
	}
	s.loadDb()
	go func() {
		for {
			if u, err := uuid.NewV4(); err == nil {
				key := strings.Replace(u.String(), "-", "", 10)
				s.RLock()
				_, ok := s.items[key]
				s.RUnlock()
				if !ok {
					s.keys <- key
					s.Lock()
					s.items[key] = fileDescriptor{Path: key}
					s.Unlock()
				}
			}
		}
	}()
	return nil
}

func NewStore(path string) (Store, error) {
	store := &fileStore{path: path, keys: make(chan string), items: make(map[string]fileDescriptor)}
	if err := store.init(); err != nil {
		return nil, err
	}
	return store, nil
}
