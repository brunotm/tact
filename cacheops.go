package tact

import (
	"fmt"
	"time"

	"github.com/brunotm/kvs"
	"github.com/brunotm/rexon"
)

const (
	cachePrefix = "cache"
)

// GetCache returns a cached or new map[keyfield]value for the given collector
func getCache(session *session, ttl time.Duration, collname string, keyFields []string) (cache map[string][]byte, err error) {
	collector := Registry.Get(collname)

	// Return if we get an error from the underlying store other than ErrNotFound
	_, err = Store.Get(cachePrefix, collector.Name, session.node.HostName, KeyLastTimestamp)
	if err != nil && err != kvs.ErrNotFound {
		return nil, err
	}

	// If not cached run collector
	if err == kvs.ErrNotFound {
		cache, err = cacheRun(session, collector, keyFields, ttl)
		if err != nil {
			return nil, err
		}
		return cache, nil
	}

	// If everything still valid load cache from the store
	events, err := Store.GetTree(cachePrefix, collector.Name, session.node.HostName)
	if err != nil {
		return nil, err
	}

	// Build index with given keys
	cache = make(map[string][]byte)
	for i := range events {
		for kf := range keyFields {
			if key, err := rexon.JSONGetUnsafeString(events[i].Value, keyFields[kf]); err == nil {
				cache[key] = events[i].Value
			}
		}
	}
	return cache, err
}

func cacheRun(sess *session, collector *Collector, keyFields []string, ttl time.Duration) (cache map[string][]byte, err error) {
	if len(keyFields) == 0 {
		return nil, fmt.Errorf("empty key fields for caching")
	}

	// Create the cache object
	cache = make(map[string][]byte)

	wchan := make(chan []byte)
	go func() {
		collector.Start(sess.Child(collector.Name), wchan)
		close(wchan)
	}()

	batch := Store.NewBatch()
	batch.SetWithTTL(nil, ttl, cachePrefix, collector.Name, sess.node.HostName, KeyLastTimestamp)

	for event := range wchan {
		for i := range keyFields {
			key, err := rexon.JSONGetUnsafeString(event, keyFields[i])
			if err != nil {
				continue
			}
			cache[key] = event

			// Use the 1st key field to keep in Store.
			if i == 0 {
				batch.SetWithTTL(event, ttl, cachePrefix, collector.Name, sess.node.HostName, key)
			}
		}
	}

	if err = batch.Write(); err != nil {
		return nil, err
	}

	return cache, err
}
