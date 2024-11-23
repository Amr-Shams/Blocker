package storage

import (
	"fmt"
	"os"
	"sync"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/iterator"
)

var (
	instance *storage
	once     sync.Once
)

const (
	dbFile = "./tmp/blockchain_%s.data"
)

type storage struct {
	nodeId string
	db     *leveldb.DB
}

var _ StorageInterface = &storage{}

func getInstance(nodeId string) (*storage, error) {
	var err error
	once.Do(func() {
		instance = &storage{nodeId: nodeId}
		dbFilePath := fmt.Sprintf(dbFile, nodeId)
		instance.db, err = leveldb.OpenFile(dbFilePath, nil)
		if err != nil {
			fmt.Println("error opening db file")
			instance.db = nil
		}
	})
	return instance, err
}

func GetInstance(nodeId string) (StorageInterface, error) {
	if instance == nil || instance.nodeId != nodeId {
		return getInstance(nodeId)
	}
	return instance, nil
}

func (s *storage) DBExist() bool {
	dbFilePath := fmt.Sprintf(dbFile, s.nodeId)
	if _, err := os.Stat(dbFilePath); os.IsNotExist(err) {
		return false
	}
	return true
}

func (s *storage) NewIterator() iterator.Iterator {
	return s.db.NewIterator(nil, nil)
}

func (s *storage) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

func (s *storage) Clean() error {
	if s.db != nil {
		s.db.Close()
	}
	dbFilePath := fmt.Sprintf(dbFile, s.nodeId)
	return os.RemoveAll(dbFilePath)
}
