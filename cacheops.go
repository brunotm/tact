package tact

import (
	"fmt"
	"time"

	"github.com/brunotm/rexon"
	"github.com/brunotm/tact/proto"
	"github.com/brunotm/tact/storage"
)

var (
	cachePrefix = []byte(`cache`)
)

// GetCache returns a cached or new map[keyfield]value for the given collector
func getCache(sess *Session, ttl time.Duration, collname string, keyFields []string) (cache map[string][]byte, err error) {

	// Get last data
	data, err := sess.txn.Get(append(cachePrefix, []byte(collname+"/"+sess.node.HostName)...))

	// If not cached run collector
	if err == storage.ErrKeyNotFound {
		return cacheRun(sess, collname, keyFields, ttl)
	}

	if err != nil {
		return nil, err
	}

	cacheData := &proto.Cache{}
	err = cacheData.Unmarshal(data)
	if err != nil {
		return nil, err
	}

	// Build index with given keys
	cache = make(map[string][]byte)
	for i := range cacheData.Data {
		for kf := range keyFields {
			if key, err := rexon.JSONGetUnsafeString(cacheData.Data[i], keyFields[kf]); err == nil {
				cache[key] = cacheData.Data[i]
			} else {
				sess.LogErr("could not index: %s, with key: %s, error: %s", string(cacheData.Data[i]), keyFields[kf], err.Error())
			}
		}
	}
	return cache, err
}

func cacheRun(sess *Session, collname string, keyFields []string, ttl time.Duration) (cache map[string][]byte, err error) {
	collector := Registry.Get(collname)

	if len(keyFields) == 0 {
		return nil, fmt.Errorf("empty key fields for caching")
	}

	// Create the cache object
	cache = make(map[string][]byte)

	wchan := make(chan []byte)
	go func() {
		collector.Start(sess.child(collector.Name), wchan)
		close(wchan)
	}()

	cacheData := &proto.Cache{}
	for event := range wchan {
		for i := range keyFields {
			key, err := rexon.JSONGetUnsafeString(event, keyFields[i])
			if err != nil {
				continue
			}
			cache[key] = event
			cacheData.Data = append(cacheData.Data, event)
		}
	}

	data, err := cacheData.Marshal()
	if err != nil {
		return nil, err
	}

	return cache, sess.txn.SetWithTTL(append(cachePrefix, []byte(collname+"/"+sess.node.HostName)...), data, ttl)
}
