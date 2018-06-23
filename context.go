package tact

import (
	"context"
	"encoding/binary"
	"fmt"
	"time"

	"github.com/brunotm/tact/collector/keys"
	"github.com/brunotm/tact/js"
	"github.com/brunotm/tact/log"
	"github.com/brunotm/tact/storage"
)

var (
	keyLastTime = []byte(keys.LastRunTime)
)

// Context for running collectors and childrens
// wraps a context.Context for cancelation
type Context struct {
	name           string
	node           *Node
	ctx            context.Context
	ctxCancel      context.CancelFunc
	timeout        time.Duration
	lastRunTime    time.Time
	currentRunTime time.Time
	cache          map[string]map[string][]byte
	dataPrefix     []byte
	store          storage.Store
	txn            storage.Txn
}

// NewContext creates a new session
func NewContext(ctx context.Context, name string, node *Node, store storage.Store, ttl time.Duration) (c *Context, err error) {
	c = &Context{}
	c.name = name
	c.node = node
	c.store = store
	c.txn = store.NewTxn(true)
	c.dataPrefix = []byte(fmt.Sprintf("session/%s/%s/", name, node.HostName))
	c.cache = make(map[string]map[string][]byte)

	c.timeout = ttl
	c.ctx, c.ctxCancel = context.WithTimeout(ctx, c.timeout)

	c.loadLastTime()
	c.currentRunTime = time.Now()

	return c, nil
}

// Child creates a new session within the current session context sharing the same cache
func (c *Context) child(name string) (child *Context, err error) {
	if child, err = NewContext(c.ctx, name, c.node, c.store, c.timeout); err != nil {
		return nil, err
	}
	child.cache = c.cache
	return child, nil
}

// Get value for the given key
func (c *Context) Get(key []byte) (value []byte, err error) {
	return c.txn.Get(append(c.dataPrefix, key...))
}

// GetTree for the given prefix
func (c *Context) GetTree(prefix []byte) (entries []storage.Entry, err error) {
	return c.txn.GetTree(append(c.dataPrefix, prefix...))
}

// Set value for the given key
func (c *Context) Set(key, value []byte) (err error) {
	return c.txn.Set(append(c.dataPrefix, key...), value)
}

// SetWithTTL value for the given key
func (c *Context) SetWithTTL(key, value []byte, ttl time.Duration) (err error) {
	return c.txn.SetWithTTL(append(c.dataPrefix, key...), value, ttl)
}

// Delete the given key
func (c *Context) Delete(key []byte) (err error) {
	return c.txn.Delete(append(c.dataPrefix, key...))
}

// DeleteTree for the given prefix
func (c *Context) DeleteTree(prefix []byte) (err error) {
	return c.txn.DeleteTree(append(c.dataPrefix, prefix...))
}

// Node returns this session node
func (c *Context) Node() *Node {
	return c.node
}

// Name returns this session name
func (c *Context) Name() string {
	return c.name
}

// Context returns this session context
func (c *Context) Context() context.Context {
	return c.ctx
}

// CurrentRunTime returns the last time this session successfully ran
func (c *Context) CurrentRunTime() time.Time {
	return c.currentRunTime
}

// LastRunTime returns the last time this session successfully ran
func (c *Context) LastRunTime() time.Time {
	return c.lastRunTime
}

// Timeout returns the session configured timeout
func (c *Context) Timeout() time.Duration {
	return c.timeout
}

// Done successfully terminates session and commit pending data
func (c *Context) done() (err error) {
	return c.close(true)
}

// Cancel session and discard pending data
func (c *Context) cancel() {
	c.close(false)
}

func (c *Context) close(ok bool) (err error) {
	if ok {
		if err = c.storeLastTime(); err != nil {
			return err
		}
		if err = c.txn.Commit(); err != nil {
			return err
		}
		c.LogDebug("commited session data")
	}
	c.ctxCancel()
	c.cache = nil
	return nil
}

func (c *Context) storeLastTime() (err error) {
	t, err := c.currentRunTime.MarshalJSON()
	if err != nil {
		return err
	}
	return c.Set(keyLastTime, t)
}

func (c *Context) loadLastTime() (err error) {
	t, err := c.Get(keyLastTime)
	if err != nil && err != storage.ErrKeyNotFound {
		return err
	}
	return c.lastRunTime.UnmarshalJSON(t)
}

// EnrichEvent enriches and outgoing event
func (c *Context) enrichEvent(event []byte) (out []byte) {
	if !js.Has(event, keys.Time) {
		event, _ = js.Set(event, c.currentRunTime, keys.Time)
	}
	event, _ = js.Set(event, c.name, keys.Metric)
	event, _ = js.Set(event, c.node.HostName, keys.Host)
	return event
}

// LogInfo the given string format with given arguments
func (c *Context) LogInfo(message string, keysAndValues ...interface{}) {
	keysAndValues = append(keysAndValues, keys.Node, c.node.HostName, keys.Collector, c.name)
	log.Info(message, keysAndValues...)

}

// LogWarn the given string format with given arguments
func (c *Context) LogWarn(message string, keysAndValues ...interface{}) {
	keysAndValues = append(keysAndValues, keys.Node, c.node.HostName, keys.Collector, c.name)
	log.Warn(message, keysAndValues...)
}

// LogError the given string format with given arguments
func (c *Context) LogError(message string, keysAndValues ...interface{}) {
	keysAndValues = append(keysAndValues, keys.Node, c.node.HostName, keys.Collector, c.name)
	log.Error(message, keysAndValues...)
}

// LogDebug the given string format with given arguments
func (c *Context) LogDebug(message string, keysAndValues ...interface{}) {
	keysAndValues = append(keysAndValues, keys.Node, c.node.HostName, keys.Collector, c.name)
	log.Debug(message, keysAndValues...)
}

func uint64Bytes(v uint64) (b []byte) {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, v)
	return buf
}

func bytesUint64(b []byte) (v uint64) {
	return binary.LittleEndian.Uint64(b)
}
