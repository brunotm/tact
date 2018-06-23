package tact

import (
	"fmt"
	"time"

	"github.com/brunotm/tact/js"
	"github.com/brunotm/tact/proto"
	"github.com/brunotm/tact/storage"
)

var (
	cachePrefix = []byte(`cache`)
)

// GetCache returns a cached or new map[keyfield]value for the given collector
func getCache(ctx *Context, ttl time.Duration, collname string, keyFields []string) (cache map[string][]byte, err error) {

	// Get last data
	data, err := ctx.txn.Get(append(cachePrefix, []byte(collname+"/"+ctx.node.HostName)...))

	// Run collector if not cached
	if err == storage.ErrKeyNotFound {
		return cacheRun(ctx, collname, keyFields, ttl)
	}

	if err != nil {
		return nil, err
	}

	cacheData := &proto.Cache{}
	err = cacheData.Unmarshal(data)
	if err != nil {
		return nil, err
	}

	// Index with given keys
	cache = make(map[string][]byte)
	for i := range cacheData.Data {
		for kf := range keyFields {
			if key, err := js.GetUnsafeString(cacheData.Data[i], keyFields[kf]); err == nil {
				cache[key] = cacheData.Data[i]
			} else {
				ctx.LogError("could not index", "event", string(cacheData.Data[i]), "key", keyFields[kf], "error", err)
			}
		}
	}
	return cache, err
}

func cacheRun(ctx *Context, collname string, keyFields []string, ttl time.Duration) (cache map[string][]byte, err error) {
	collector := Registry.Get(collname)

	if len(keyFields) == 0 {
		return nil, fmt.Errorf("empty key fields for caching")
	}

	// Create the cache object
	cache = make(map[string][]byte)

	wchan := make(chan []byte)
	go func() {
		child, err := ctx.child(collector.Name)
		if err != nil {
			ctx.LogError("creating child session", "error", err)
			return
		}
		collector.Start(child, wchan)
		close(wchan)
	}()

	cacheData := &proto.Cache{}
	for event := range wchan {
		for i := range keyFields {
			key, err := js.GetUnsafeString(event, keyFields[i])
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

	return cache, ctx.txn.SetWithTTL(append(cachePrefix, []byte(collname+"/"+ctx.node.HostName)...), data, ttl)
}
