package tact

import (
	"time"

	"github.com/brunotm/rexon"
	"github.com/brunotm/tact/storage"
)

var (
	deltaPrefix = []byte(`delta`)
)

// Blacklist type
type Blacklist map[string]struct{}

// Add blackilists a given key
func (b Blacklist) Add(key string) {
	b[key] = struct{}{}
}

// DeltaOps type
type DeltaOps struct {
	KeyField      string        // Field to use as unique key
	Rate          bool          // Wetheter or not to do rate calculations over time delta
	TTL           time.Duration // TTL of cached data
	Blacklist     Blacklist     // Fields to ignore in delta, can be nil
	RateBlacklist Blacklist     // Fields to exclude from rate calculations, can be nil
}

// EventOps type
type EventOps struct {
	Round        int                        // The precision for float fields
	FieldTypes   map[string]rexon.ValueType // The Fields:Type for conversion
	FieldRenames map[string]string          // The Fields:Name for renaming
	Delta        *DeltaOps
}

func (eo *EventOps) process(sess *Session, event []byte) []byte {
	var err error
	// Perform any specified field ops
	if eo.FieldTypes != nil {
		result := rexon.ParseJsonValues(event, eo.FieldTypes, eo.Round)
		if result.Errors != nil {
			for err := range result.Errors {
				sess.LogErr(result.Errors[err].Error())
			}
			return nil
		}
		event = result.Data
	}

	// Perform any specified delta ops
	if eo.Delta != nil {
		event, err = eo.eventDelta(sess, event)
		if err != nil {
			sess.LogErr("%s calculating deltas for event: %s", err.Error(), event)
			return nil
		}
	}

	return event
}

// eventDelta perform delta and rate calculation
func (eo *EventOps) eventDelta(sess *Session, event []byte) ([]byte, error) {

	// Get any specified event unique attribute key, empty string otherwise
	keyVal, _ := rexon.JSONGetUnsafeString(event, eo.Delta.KeyField)

	// Set the timestamp on the current event for caching
	event, err := rexon.JSONSet(event, sess.timeCurrent, KeyTimeStamp)
	if err != nil {
		return nil, err
	}

	// Get previous event for delta.
	// If we can't find a existing event, store the current event and return
	key := append(deltaPrefix, []byte(sess.name+"/"+sess.node.HostName+"/"+keyVal)...)
	previous, err := sess.txn.Get(key)

	if err != nil {
		if err == storage.ErrKeyNotFound {
			return nil, sess.txn.SetWithTTL(key, event, eo.Delta.TTL)
		}
		return nil, err
	}

	// Store the current event
	err = sess.txn.SetWithTTL(key, event, eo.Delta.TTL)
	if err != nil {
		return nil, err
	}

	// Parse the current and previous events, and timestamps for calculations
	previousTimestamp, _ := rexon.JSONGetInt(previous, KeyTimeStamp)
	timeDelta := sess.timeCurrent - previousTimestamp

	// Loop over the event fields and perform the operations specified for each one.
	// if we get an error calculating stop return an error event to the stream
	err = rexon.JSONForEach(event, func(key string, value []byte, tp rexon.ValueType) error {
		if key == KeyTimeStamp {
			return nil
		}

		// Skip if the current field is blacklisted
		if eo.Delta.Blacklist != nil {
			if _, ok := eo.Delta.Blacklist[key]; ok {
				return nil
			}
		}

		// Get the previous value for calculation and check its type
		// If its not a numeric type stop the iteration
		previousValue, err := rexon.JSONGetFloat(previous, key)
		if err != nil {
			return err
		}

		currentValue, err := rexon.JSONGetFloat(value)
		if err != nil {
			return err
		}

		// Get the delta for this key
		newValue := currentValue - previousValue

		// Check if have any rate calculation to do and if current field isn't blacklisted
		// If not calculate and round to the precision on FieldOps.Round
		if !eo.Delta.Rate {
			event, err = rexon.JSONSet(event, newValue, key)
			return err
		}

		if _, ok := eo.Delta.RateBlacklist[key]; ok {
			event, err = rexon.JSONSet(event, newValue, key)
			return err
		}

		event, err = rexon.JSONSet(event, rexon.Round(newValue/float64(timeDelta), eo.Round), key)
		return err
	})

	return event, err
}
