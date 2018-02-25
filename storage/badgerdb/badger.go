package badgerdb

import (
	"os"
	"time"

	"github.com/brunotm/tact/storage"

	"github.com/dgraph-io/badger"
)

const (
	gcDiscardRatio = 0.5
)

var (
	// Check if Store satisfies kvs.Store interface.
	_ storage.Store = (*Store)(nil)
	// Check if Store satisfies kvs.Store interface.
	_ storage.Txn = (*Txn)(nil)
)

// Store type
type Store struct {
	db   *badger.DB
	path string
}

// Open or creates a store
func Open(path string) (store *Store, err error) {
	if err = os.MkdirAll(path, 0755); err != nil {
		return nil, err
	}
	opts := badger.DefaultOptions
	opts.Dir = path
	opts.ValueDir = path
	opts.MaxTableSize = 8 << 20      // def 64 << 20
	opts.NumMemtables = 3            // def 5
	opts.NumLevelZeroTables = 3      // def 5
	opts.NumLevelZeroTablesStall = 5 // def 10

	var db *badger.DB
	db, err = badger.Open(opts)
	if err != nil {
		return store, err
	}

	store = &Store{}
	store.db = db
	store.path = path

	return store, nil

}

// Close the current Store
func (s *Store) Close() (err error) {
	return s.db.Close()
}

// Remove the current Store
func (s *Store) Remove() (err error) {
	s.db.Close()
	return os.RemoveAll(s.path)
}

// RunGC garbage collect the undelying DB
func (s *Store) RunGC() (err error) {
	if err = s.db.PurgeOlderVersions(); err != nil {
		return err
	}

	if err = s.db.RunValueLogGC(gcDiscardRatio); err == badger.ErrNoRewrite {
		return nil
	}
	return err
}

// NewTxn creates a rw/ro transaction
func (s *Store) NewTxn(update bool) (txn storage.Txn) {
	return &Txn{s.db.NewTransaction(update)}
}

// Txn transaction
type Txn struct {
	txn *badger.Txn
}

// Discard this transaction
func (t *Txn) Discard() {
	t.txn.Discard()
}

// Commit this transaction
func (t *Txn) Commit() (err error) {
	return t.txn.Commit(nil)
}

// Get value for the given key
func (t *Txn) Get(key []byte) (value []byte, err error) {
	item, err := t.txn.Get(key)
	if err == badger.ErrKeyNotFound {
		return nil, storage.ErrKeyNotFound
	}

	if err != nil {
		return nil, err
	}

	if value, err = item.Value(); err != nil {
		return nil, err
	}
	return storage.SnappyDecode(value)
}

// GetTree for the given prefix
func (t *Txn) GetTree(prefix []byte) (entries []storage.Entry, err error) {
	var entry storage.Entry
	var value []byte

	it := t.txn.NewIterator(badger.DefaultIteratorOptions)
	defer it.Close()

	for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
		item := it.Item()
		if value, err = item.Value(); err != nil {
			return nil, err
		}

		if entry.Value, err = storage.SnappyDecode(value); err != nil {
			return nil, err
		}

		entry.Key = storage.CopyBytes(item.Key())
		entries = append(entries, entry)
	}
	return entries, nil
}

// Set value for the given key
func (t *Txn) Set(key, value []byte) (err error) {
	return t.txn.Set(key, storage.SnappyEncode(value))
}

// SetWithTTL value for the given key
func (t *Txn) SetWithTTL(key, value []byte, ttl time.Duration) (err error) {
	return t.txn.SetWithTTL(key, storage.SnappyEncode(value), ttl)
}

// Delete the given key
func (t *Txn) Delete(key []byte) (err error) {
	return t.txn.Delete(key)
}

// DeleteTree for the given prefix
func (t *Txn) DeleteTree(prefix []byte) (err error) {
	opts := badger.DefaultIteratorOptions
	opts.PrefetchValues = false

	it := t.txn.NewIterator(badger.DefaultIteratorOptions)
	defer it.Close()

	for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
		item := it.Item()
		if err = t.txn.Delete(item.Key()); err != nil {
			return err
		}
	}
	return nil
}
