package tact

import (
	"context"
	"fmt"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/brunotm/rexon"
)

// Session holds the context and session data
// for running collectors and childrens
type Session struct {
	name        string
	node        *Node
	ctx         context.Context
	ctxCancel   context.CancelFunc
	timeCurrent int64
	timeLast    int64
	timeout     time.Duration
	cache       map[string]map[string][]byte
}

// NewSession creates a new session
func NewSession(ctx context.Context, name string, node *Node, ttl time.Duration) (s *Session) {
	s = &Session{}
	s.name = name
	s.ctx = ctx
	s.timeout = ttl
	s.node = node
	return s
}

// Start session setting up last time, cache storage and cancellation
func (s *Session) Start() {
	// TODO: refactor to error out
	s.loadLastTime()
	s.cache = make(map[string]map[string][]byte)
	s.timeCurrent = time.Now().Unix()

	if s.timeout > 0 {
		s.ctx, s.ctxCancel = context.WithTimeout(s.ctx, s.timeout)
	} else {
		s.ctx, s.ctxCancel = context.WithCancel(s.ctx)
	}
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
func (s *Session) child(name string) *Session {
	return NewSession(s.ctx, name, s.node, s.timeout)
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

func (s *Session) close(ok bool) {
	if ok {
		s.storeLastTime()
	}
	s.ctxCancel()
	s.timeCurrent = 0
	s.timeLast = 0
	s.cache = nil
}

func (s *Session) storeLastTime() {
	lb := uint64Bytes(uint64(s.timeCurrent))
	if err := Store.Set(lb, s.name, s.node.HostName, KeyLastTimestamp); err != nil {
		s.LogErr("store last time: %s", err.Error())
	}
}

func (s *Session) loadLastTime() {
	lb, err := Store.Get(s.name, s.node.HostName, KeyLastTimestamp)
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
