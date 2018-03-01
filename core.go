package tact

import (
	"sync"

	log "github.com/Sirupsen/logrus"
	"github.com/brunotm/tact/storage"
	"github.com/brunotm/tact/storage/badgerdb"
)

var (
	// workingPath, _ = filepath.Abs(filepath.Dir(os.Args[0]))

	// Registry default collector registry
	Registry *registry
	// Store default persistence store
	Store storage.Store
)

// init the core
func init() {
	Registry = &registry{
		mtx:        &sync.RWMutex{},
		collectors: map[string]*Collector{},
		groups:     map[string][]*Collector{},
	}
}

// Init initializes core structures
func Init(path string) {
	var err error
	Store, err = badgerdb.Open(path, true)
	if err != nil {
		panic(err)
	}
}

// Close shutdown and stops the core
func Close() {
	if err := Store.Close(); err != nil {
		log.WithFields(
			log.Fields{
				"error": err.Error(),
			},
		).Info("Error closing KVStore")
	}
}
