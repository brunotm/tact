package tact

import (
	"context"
	"fmt"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/brunotm/rexon"
	"github.com/brunotm/tact/storage"
)

var (
	keyLastTime = []byte("last_timestamp")
)

// Session holds the context and session data
// for running collectors and childrens
type Session struct {
	name        string
	node        *Node
	ctx         context.Context
	ctxParent   context.Context
	ctxCancel   context.CancelFunc
	timeout     time.Duration
	timeCurrent int64
	timeLast    int64
	cache       map[string]map[string][]byte
	dataPath    []byte
	store       storage.Store
	txn         storage.Txn
}

// NewSession creates a new session
func NewSession(ctx context.Context, name string, node *Node, store storage.Store, ttl time.Duration) (s *Session) {
	s = &Session{}
	s.name = name
	s.ctxParent = ctx
	s.timeout = ttl
	s.node = node
	s.store = store
	s.dataPath = []byte(fmt.Sprintf("session/%s/%s/", name, node.HostName))
	return s
}

// Start session setting up last time, cache storage and cancellation
func (s *Session) Start() {
	// TODO: refactor to error out
	s.txn = s.store.NewTxn(true)
	s.loadLastTime()
	s.timeCurrent = time.Now().Unix()

	// Check if we're not sharing the underlying
	// cache with child sessions
	if s.cache == nil {
		s.cache = make(map[string]map[string][]byte)
	}

	if s.timeout > 0 {
		s.ctx, s.ctxCancel = context.WithTimeout(s.ctxParent, s.timeout)
	} else {
		s.ctx, s.ctxCancel = context.WithCancel(s.ctxParent)
	}

	// Ensure we discard our storage transaction
	go func() {
		<-s.ctx.Done()
		s.txn.Discard()
		s.LogDebug("session context canceled and Txn discarded")
	}()

	s.LogDebug("session context and storage Txn created")

}

// Get value for the given key
func (s *Session) Get(key []byte) (value []byte, err error) {
	return s.txn.Get(append(s.dataPath, key...))
}

// GetTree for the given prefix
func (s *Session) GetTree(prefix []byte) (entries []storage.Entry, err error) {
	return s.txn.GetTree(append(s.dataPath, prefix...))
}

// Set value for the given key
func (s *Session) Set(key, value []byte) (err error) {
	return s.txn.Set(append(s.dataPath, key...), value)
}

// SetWithTTL value for the given key
func (s *Session) SetWithTTL(key, value []byte, ttl time.Duration) (err error) {
	return s.txn.SetWithTTL(append(s.dataPath, key...), value, ttl)
}

// Delete the given key
func (s *Session) Delete(key []byte) (err error) {
	return s.txn.Delete(append(s.dataPath, key...))
}

// DeleteTree for the given prefix
func (s *Session) DeleteTree(prefix []byte) (err error) {
	return s.txn.DeleteTree(append(s.dataPath, prefix...))
}

// Node returns this session node
func (s *Session) Node() *Node {
	return s.node
}

// Name returns this session name
func (s *Session) Name() string {
	return s.name
}

// Context returns this session context
func (s *Session) Context() context.Context {
	return s.ctx
}

// LastTime returns the last time this session successfully ran
func (s *Session) LastTime() int64 {
	return s.timeLast
}

// done successfully terminates session
func (s *Session) done() {
	s.close(true)
}

// cancel session
func (s *Session) cancel() {
	s.close(false)
}

// child creates a new session within the current session context
func (s *Session) child(name string) (child *Session) {
	child = NewSession(s.ctx, name, s.node, s.store, s.timeout)
	child.cache = s.cache
	return child
}

func (s *Session) close(ok bool) {
	if ok {
		s.LogDebug("commiting session data")
		s.storeLastTime()
		if err := s.txn.Commit(); err != nil {
			s.LogErr("%s writing session data", err.Error())
		}
	}
	s.LogDebug("cancel session context and reset state")
	s.ctxCancel()
	s.timeCurrent = 0
	s.timeLast = 0
	s.cache = nil
}

func (s *Session) storeLastTime() {
	lb := uint64Bytes(uint64(s.timeCurrent))
	if err := s.Set(keyLastTime, lb); err != nil {
		s.LogErr("storing last time: %s", err.Error())
	}
}

func (s *Session) loadLastTime() {
	lb, err := s.Get(keyLastTime)
	if err != nil {
		return
	}
	s.timeLast = int64(bytesUint64(lb))
}

// EnrichEvent enriches and outgoing event
func (s *Session) enrichEvent(event []byte) []byte {
	if !rexon.JSONExists(event, KeyTimeStamp) {
		event, _ = rexon.JSONSet(event, s.timeCurrent, KeyTimeStamp)
	}
	event, _ = rexon.JSONSet(event, s.name, KeyMetric)
	event, _ = rexon.JSONSet(event, s.node.HostName, KeyHostName)
	return event
}

// LogInfo the given string format with given arguments
func (s *Session) LogInfo(message string, args ...interface{}) {
	log.WithFields(s.getLogFields()).Info(logMessage(message, args))
}

// LogWarn the given string format with given arguments
func (s *Session) LogWarn(message string, args ...interface{}) {
	log.WithFields(s.getLogFields()).Warn(logMessage(message, args))
}

// LogErr the given string format with given arguments
func (s *Session) LogErr(message string, args ...interface{}) {
	log.WithFields(s.getLogFields()).Error(logMessage(message, args))
}

// LogDebug the given string format with given arguments
func (s *Session) LogDebug(message string, args ...interface{}) {
	log.WithFields(s.getLogFields()).Debug(logMessage(message, args))
}

// return a log.Fields
func (s *Session) getLogFields() log.Fields {
	return log.Fields{
		KeyNode:      s.node.HostName,
		KeyCollector: s.name,
	}
}

// logMessage
func logMessage(message string, args []interface{}) string {
	return fmt.Sprintf(message, args...)
}
