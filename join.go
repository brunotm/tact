package tact

import (
	"fmt"
	"time"

	"github.com/brunotm/tact/js"
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
func (j *Join) Process(ctx *Context, event []byte) (joined []byte, ok bool) {
	// Try to join on each specified field until matched
	for field := range j.JoinFields {
		if event, ok := j.join(ctx, event, j.JoinFields[field]); ok {
			return event, ok
		}
	}
	// Return unjoined
	return event, false
}

func (j *Join) join(ctx *Context, event []byte, joinField string) (joined []byte, ok bool) {
	eventKey, _ := js.GetUnsafeString(event, joinField)
	cached, ok := ctx.cache[j.Name][eventKey]
	if ok {
		// Include specified fields
		for _, field := range j.IncludeFields {
			value, _ := js.GetValue(cached, field)
			event, _ = js.Set(event, value, field)
		}
		return event, ok
	}
	// Return unjoined
	return event, false
}

func (j *Join) loadData(ctx *Context) (err error) {
	cache, err := getCache(ctx, j.TTL, j.Name, j.JoinOnFields)
	if err != nil {
		return fmt.Errorf("cache load for %s error: %s", j.Name, err.Error())
	}
	ctx.cache[j.Name] = cache
	return nil
}
