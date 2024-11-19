package storage

import (
	"encoding/binary"

	"github.com/Amr-Shams/Blocker/util"
	"github.com/syndtr/goleveldb/leveldb"
)

func (s *storage) SaveLastHash(hash []byte) error {
	return s.db.Put([]byte("lh"), hash, nil)
}

func (s *storage) SaveLastTimeUpdate(TimeStamp int64) error {
	return s.db.Put([]byte("ltu"), util.IntToHex(TimeStamp), nil)
}

func (s *storage) SaveHeight(height int) error {
	return s.db.Put([]byte("h"), util.IntToHex(int64(height)), nil)
}

func (s *storage) GetLastHash() ([]byte, error) {
	return s.db.Get([]byte("lh"), nil)
}

func (s *storage) GetLastTimeUpdate() (int64, error) {
	data, err := s.db.Get([]byte("ltu"), nil)
	if err != nil {
		return 0, err
	}
	return int64(binary.LittleEndian.Uint64(data)), nil
}

func (s *storage) GetHeight() (int, error) {
	data, err := s.db.Get([]byte("h"), nil)
	if err != nil {
		return 0, err
	}
	return util.BytesToInt(data), nil
}

func (s *storage) SaveBlock(key []byte, value []byte) error {
	return s.db.Put(key, value, nil)
}

func (s *storage) GetBlock(key []byte) ([]byte, error) {
	return s.db.Get(key, nil)
}

func (s *storage) SaveBatch(batch *leveldb.Batch) error {
	return s.db.Write(batch, nil)
}

func (s *storage) DeleteBlock(key []byte) error {
	return s.db.Delete(key, nil)
}
