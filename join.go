package tact

import (
	"fmt"
	"time"

	"github.com/brunotm/rexon"
)

// Join type
type Join struct {
	Name          string        // Collector name eg. `/aix/config/lvm`
	TTL           time.Duration // TTL when using cache
	JoinFields    []string      // Field names for possible matches to join, it will successfully return on first match
	JoinOnFields  []string      // Field name from the events of called Collector to join on
	IncludeFields []string      // Fields to include from the events of called collector
}

// Process joins for the given event
func (j *Join) Process(session *session, event []byte) (joined []byte, ok bool) {
	// Try to join on each specified field until matched
	for field := range j.JoinFields {
		if event, ok := j.join(session, event, j.JoinFields[field]); ok {
			return event, ok
		}
	}
	// Return unjoined
	return event, false
}

func (j *Join) join(session *session, event []byte, joinField string) (joined []byte, ok bool) {
	eventKey, _ := rexon.JSONGetUnsafeString(event, joinField)
	cached, ok := session.cache[j.Name][eventKey]
	if ok {
		// Include specified fields
		for _, field := range j.IncludeFields {
			value, _, _ := rexon.JSONGetValue(cached, field)
			event, _ = rexon.JSONSet(event, value, field)
		}
		return event, ok
	}
	// Return unjoined
	return event, false
}

func (j *Join) loadData(session *session) (err error) {
	// TODO: make loading indempotent
	cache, err := getCache(session, j.TTL, j.Name, j.JoinOnFields)
	if err != nil {
		return fmt.Errorf("cache load for %s error: %s", j.Name, err.Error())
	}
	session.cache[j.Name] = cache
	return nil
}
