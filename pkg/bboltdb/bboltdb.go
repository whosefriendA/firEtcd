// Package bboltdb provides a bbolt-backed implementation for the storage interface.
// It simulates an in-memory database using a temporary file.
package bboltdb

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/whosefriendA/firEtcd/common"
	"go.etcd.io/bbolt"
)

var defaultBucket = []byte("kv")

// BoltDB implements the common.DB interface using bbolt.
// It uses a temporary file on disk to simulate in-memory storage,
// which is automatically cleaned up on Close().
type BoltDB struct {
	db   *bbolt.DB
	path string // Store the path for cleanup
}

// Compile-time check to ensure BoltDB implements the common.DB interface.
var _ common.DB = &BoltDB{}

// NewDB creates a new "in-memory" bbolt database.
func NewDB() *BoltDB {
	// Create a temporary file to act as the database.
	tmpfile, err := ioutil.TempFile("", "bbolt-*.db")
	if err != nil {
		panic(fmt.Sprintf("failed to create temp file for bbolt: %v", err))
	}
	path := tmpfile.Name()
	tmpfile.Close() // Close the file so bbolt can open it.

	db, err := bbolt.Open(path, 0600, &bbolt.Options{Timeout: 2 * time.Second})
	if err != nil {
		panic(fmt.Sprintf("failed to open bbolt db at %s: %v", path, err))
	}

	// Ensure the default bucket exists.
	err = db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(defaultBucket)
		return err
	})
	if err != nil {
		panic(fmt.Sprintf("failed to create default bucket: %v", err))
	}

	return &BoltDB{db: db, path: path}
}

// Close closes the database and removes the temporary file.
func (b *BoltDB) Close() error {
	err := b.db.Close()
	if err != nil {
		// Still try to remove the temp file even if close fails.
		os.Remove(b.path)
		return err
	}
	return os.Remove(b.path)
}

func (b *BoltDB) Del(key string) error {
	return b.db.Update(func(tx *bbolt.Tx) error {
		return tx.Bucket(defaultBucket).Delete([]byte(key))
	})
}

func (b *BoltDB) DelWithPrefix(prefix string) error {
	return b.db.Update(func(tx *bbolt.Tx) error {
		c := tx.Bucket(defaultBucket).Cursor()
		prefixBytes := []byte(prefix)
		for k, _ := c.Seek(prefixBytes); k != nil && bytes.HasPrefix(k, prefixBytes); k, _ = c.Next() {
			// c.Delete() is more efficient inside a loop than bucket.Delete()
			if err := c.Delete(); err != nil {
				return err
			}
		}
		return nil
	})
}

// Get is not directly used by kvraft but implemented for interface completeness.
func (b *BoltDB) Get(key string) (ret []byte, err error) {
	err = b.db.View(func(tx *bbolt.Tx) error {
		v := tx.Bucket(defaultBucket).Get([]byte(key))
		if v == nil {
			return common.ErrNotFound
		}
		// We must copy the value, as it's only valid for the transaction lifetime.
		ret = make([]byte, len(v))
		copy(ret, v)
		return nil
	})
	return
}

// GetWithPrefix is not directly used by kvraft but implemented for interface completeness.
func (b *BoltDB) GetWithPrefix(prefix string) (ret [][]byte, err error) {
	ret = make([][]byte, 0, 1)
	err = b.db.View(func(tx *bbolt.Tx) error {
		c := tx.Bucket(defaultBucket).Cursor()
		prefixBytes := []byte(prefix)
		for k, v := c.Seek(prefixBytes); k != nil && bytes.HasPrefix(k, prefixBytes); k, v = c.Next() {
			valCopy := make([]byte, len(v))
			copy(valCopy, v)
			ret = append(ret, valCopy)
		}
		return nil
	})
	return
}

func (b *BoltDB) GetEntry(key string) (ret common.Entry, err error) {
	err = b.db.View(func(tx *bbolt.Tx) error {
		v := tx.Bucket(defaultBucket).Get([]byte(key))
		if v == nil {
			return common.ErrNotFound
		}
		if err := gob.NewDecoder(bytes.NewReader(v)).Decode(&ret); err != nil {
			return err
		}
		return nil
	})
	return
}

func (b *BoltDB) GetEntryWithPrefix(prefix string) (ret []common.Entry, err error) {
	ret = make([]common.Entry, 0, 1)
	err = b.db.View(func(tx *bbolt.Tx) error {
		c := tx.Bucket(defaultBucket).Cursor()
		prefixBytes := []byte(prefix)
		for k, v := c.Seek(prefixBytes); k != nil && bytes.HasPrefix(k, prefixBytes); k, v = c.Next() {
			var entry common.Entry
			if gob.NewDecoder(bytes.NewReader(v)).Decode(&entry) == nil {
				ret = append(ret, entry)
			}
		}
		return nil
	})
	return
}

func (b *BoltDB) GetPairsWithPrefix(prefix string) (ret []common.Pair, err error) {
	ret = make([]common.Pair, 0, 1)
	err = b.db.View(func(tx *bbolt.Tx) error {
		c := tx.Bucket(defaultBucket).Cursor()
		prefixBytes := []byte(prefix)
		for k, v := c.Seek(prefixBytes); k != nil && bytes.HasPrefix(k, prefixBytes); k, v = c.Next() {
			var entry common.Entry
			if gob.NewDecoder(bytes.NewReader(v)).Decode(&entry) == nil {
				keyCopy := make([]byte, len(k))
				copy(keyCopy, k)
				ret = append(ret, common.Pair{Key: string(keyCopy), Entry: entry})
			}
		}
		return nil
	})
	return
}

// Put is not directly used by kvraft but implemented for interface completeness.
func (b *BoltDB) Put(key string, value []byte, DeadTime int64) error {
	entry := common.Entry{
		Value:    value,
		DeadTime: DeadTime,
	}
	return b.PutEntry(key, entry)
}

func (b *BoltDB) PutEntry(key string, entry common.Entry) error {
	return b.db.Update(func(tx *bbolt.Tx) error {
		var buf bytes.Buffer
		if err := gob.NewEncoder(&buf).Encode(entry); err != nil {
			return err
		}
		return tx.Bucket(defaultBucket).Put([]byte(key), buf.Bytes())
	})
}

func (b *BoltDB) Keys(pageSize, pageIndex int) (ret []common.Pair, err error) {
	ret = make([]common.Pair, 0, pageSize)
	err = b.db.View(func(tx *bbolt.Tx) error {
		c := tx.Bucket(defaultBucket).Cursor()
		skip := pageIndex * pageSize
		count := 0

		// Move cursor to the first element of the page
		k, v := c.First()
		for i := 0; i < skip && k != nil; i++ {
			k, v = c.Next()
		}

		// Read the page
		for ; k != nil && count < pageSize; k, v = c.Next() {
			var entry common.Entry
			if gob.NewDecoder(bytes.NewReader(v)).Decode(&entry) == nil {
				keyCopy := make([]byte, len(k))
				copy(keyCopy, k)
				entry.Value = nil // As per original buntdb implementation
				ret = append(ret, common.Pair{Key: string(keyCopy), Entry: entry})
				count++
			}
		}
		return nil
	})
	return
}

func (b *BoltDB) KVs(pageSize, pageIndex int) (ret []common.Pair, err error) {
	ret = make([]common.Pair, 0, pageSize)
	err = b.db.View(func(tx *bbolt.Tx) error {
		c := tx.Bucket(defaultBucket).Cursor()
		skip := pageIndex * pageSize
		count := 0

		k, v := c.First()
		for i := 0; i < skip && k != nil; i++ {
			k, v = c.Next()
		}

		for ; k != nil && count < pageSize; k, v = c.Next() {
			var entry common.Entry
			if gob.NewDecoder(bytes.NewReader(v)).Decode(&entry) == nil {
				keyCopy := make([]byte, len(k))
				copy(keyCopy, k)
				ret = append(ret, common.Pair{Key: string(keyCopy), Entry: entry})
				count++
			}
		}
		return nil
	})
	return
}

func (b *BoltDB) SnapshotData() ([]byte, error) {
	var buf bytes.Buffer
	err := b.db.View(func(tx *bbolt.Tx) error {
		_, err := tx.WriteTo(&buf)
		return err
	})
	return buf.Bytes(), err
}

func (b *BoltDB) InstallSnapshotData(data []byte) error {
	// This operation replaces the current database with the snapshot.
	if err := b.db.Close(); err != nil {
		return err
	}

	// Get the file path
	dbPath := b.db.Path()

	// Write the snapshot data to the file, overwriting it.
	// This assumes the snapshot data is a valid bbolt database file.
	file, err := os.OpenFile(dbPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	if _, err := file.Write(data); err != nil {
		file.Close()
		return err
	}
	file.Close()

	// Re-open the database
	newDB, err := bbolt.Open(dbPath, 0600, &bbolt.Options{Timeout: 2 * time.Second})
	if err != nil {
		return err
	}
	b.db = newDB
	return nil
}
