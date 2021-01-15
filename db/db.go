package db

import (
	"fmt"
	"github.com/boltdb/bolt"
	"path"
	"strings"
	"sync"
)

const MAX_NUM = 1000

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

func (w *BoltDB) PutRetry(k []byte) error {
	w.rwLock.Lock()
	defer w.rwLock.Unlock()

	return w.db.Update(func(btx *bolt.Tx) error {
		bucket := btx.Bucket(BKTRetry)
		err := bucket.Put(k, []byte{0x00})
		if err != nil {
			return err
		}

		return nil
	})
}

func (w *BoltDB) GetAllRetry() ([][]byte, error) {
	w.rwLock.Lock()
	defer w.rwLock.Unlock()

	retryList := make([][]byte, 0)
	err := w.db.Update(func(tx *bolt.Tx) error {
		bw := tx.Bucket(BKTRetry)
		bw.ForEach(func(k, _ []byte) error {
			_k := make([]byte, len(k))
			copy(_k, k)
			retryList = append(retryList, _k)
			if len(retryList) >= MAX_NUM {
				return fmt.Errorf("max num")
			}
			return nil
		})
		return nil
	})
	if err != nil {
		return nil, err
	}
	return retryList, nil
}
