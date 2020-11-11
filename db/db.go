package db

import (
	"github.com/boltdb/bolt"
	"path"
	"strings"
	"sync"
)

type BoltDB struct {
	rwLock   *sync.RWMutex
	db       *bolt.DB
	filePath string
}

var (
	BKTCheck = []byte("Check")
	BKTRetry = []byte("Retry")
)

func NewBoltDB(filePath string) (*BoltDB, error) {
	if !strings.Contains(filePath, ".bin") {
		filePath = path.Join(filePath, "bolt.bin")
	}

	w := new(BoltDB)
	db, err := bolt.Open(filePath, 0644, &bolt.Options{InitialMmapSize: 500000})
	if err != nil {
		return nil, err
	}

	w.db = db
	w.rwLock = new(sync.RWMutex)
	w.filePath = filePath

	// poly check
	if err = db.Update(func(btx *bolt.Tx) error {
		_, err := btx.CreateBucketIfNotExists(BKTCheck)
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}

	// poly retry
	if err = db.Update(func(btx *bolt.Tx) error {
		_, err := btx.CreateBucketIfNotExists(BKTRetry)
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return w, nil
}
