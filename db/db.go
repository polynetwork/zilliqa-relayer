package db

import (
	"encoding/binary"
	"encoding/hex"
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
	BKTCheck  = []byte("Check")
	BKTRetry  = []byte("Retry")
	BKTHeight = []byte("Height")
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

	// poly check
	if err = db.Update(func(btx *bolt.Tx) error {
		_, err := btx.CreateBucketIfNotExists(BKTHeight)
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

func (w *BoltDB) DeleteRetry(k []byte) error {
	w.rwLock.Lock()
	defer w.rwLock.Unlock()

	return w.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(BKTRetry)
		err := bucket.Delete(k)
		if err != nil {
			return err
		}
		return nil
	})
}

func (w *BoltDB) PutCheck(txHash string, v []byte) error {
	w.rwLock.Lock()
	defer w.rwLock.Unlock()

	k, err := hex.DecodeString(txHash)
	if err != nil {
		return err
	}
	return w.db.Update(func(btx *bolt.Tx) error {
		bucket := btx.Bucket(BKTCheck)
		err := bucket.Put(k, v)
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

func (w *BoltDB) UpdatePolyHeight(h uint32) error {
	w.rwLock.Lock()
	defer w.rwLock.Unlock()

	raw := make([]byte, 4)
	binary.LittleEndian.PutUint32(raw, h)

	return w.db.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(BKTHeight)
		return bkt.Put([]byte("poly_height"), raw)
	})
}

func (w *BoltDB) GetPolyHeight() uint32 {
	w.rwLock.RLock()
	defer w.rwLock.RUnlock()

	var h uint32
	_ = w.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(BKTHeight)
		raw := bkt.Get([]byte("poly_height"))
		if len(raw) == 0 {
			h = 0
			return nil
		}
		h = binary.LittleEndian.Uint32(raw)
		return nil
	})
	return h
}
