package storage

import (
	"errors"
	"time"
)

var (
	// ErrKeyNotFound error
	ErrKeyNotFound = errors.New("key not found")
)

// Entry key value
type Entry struct {
	Key, Value []byte
}

// Store interface
type Store interface {
	// Close the current Store
	Close() (err error)
	// Remove the current Store
	Remove() (err error)
	// RunGC garbage collect the undelying DB
	RunGC() (err error)
	// NewTxn creates a rw/ro transaction
	NewTxn(update bool) (txn Txn)
}

// Txn interface
type Txn interface {
	// Discard this transaction
	Discard()
	// Commit this transaction
	Commit() (err error)
	// Get value for the given key
	Get(key []byte) (value []byte, err error)
	// GetTree for the given prefix
	GetTree(prefix []byte) (entries []Entry, err error)
	// Set value for the given key
	Set(key, value []byte) (err error)
	// SetWithTTL value for the given key
	SetWithTTL(key, value []byte, ttl time.Duration) (err error)
	// Delete the given key
	Delete(key []byte) (err error)
	// DeleteTree for the given prefix
	DeleteTree(prefix []byte) (err error)
}
