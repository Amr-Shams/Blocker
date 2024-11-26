package storage

import (
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/iterator"
)

type StorageInterface interface {
	NewIterator() iterator.Iterator
	SaveLastHash(hash []byte) error
	SaveLastTimeUpdate(TimeStamp int64) error
	SaveHeight(height int64) error
	GetLastHash() ([]byte, error)
	GetLastTimeUpdate() (int64, error)
	GetHeight() (int64, error)
	SaveBlock(key []byte, value []byte) error
	SaveBatch(batch *leveldb.Batch) error
	GetBlock(key []byte) ([]byte, error)
	DeleteBlock(key []byte) error
	Close() error
	DBExist() bool
	Clean() error
}
type FileSystemInterface interface {
	WriteToFile(data []byte) error
	ReadFromFile() ([]byte, error)
	Close() error
}
