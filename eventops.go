package tact

import (
	"math"
	"time"

	"github.com/brunotm/tact/collector/keys"
	"github.com/brunotm/tact/js"

	"github.com/brunotm/rexon"
	"github.com/brunotm/tact/storage"
)

var (
	deltaPrefix = []byte(keys.Delta)
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
	Round        int               // The precision for float fields
	FieldTypes   []*rexon.Value    // The Fields:Type for conversion
	FieldRenames map[string]string // The Fields:Name for renaming
	Delta        *DeltaOps
}

func (eo *EventOps) process(ctx *Context, event []byte) (out []byte) {
	var err error

	// Perform any specified field ops
	if eo.FieldTypes != nil {
		for _, vp := range eo.FieldTypes {
			field := vp.Name()
			raw, err := js.Get(event, field)
			if err != nil {
				ctx.LogError("type conversion error fetching %s from event: %s", field, err)
				continue
			}

			value, _, err := vp.Parse(raw)
			if err != nil {
				ctx.LogError("%s type conversion error: %s", field, err)
				continue
			}

			event, err = js.Set(event, value, field)
			if err != nil {
				ctx.LogError("%s type conversion error setting: %s", field, err)
				continue
			}
		}
	}

	// Perform any specified delta ops
	if eo.Delta != nil {
		event, err = eo.eventDelta(ctx, event)
		if err != nil {
			ctx.LogError("%s calculating deltas for event: %s", err.Error(), event)
			return nil
		}
	}

	return event
}

// eventDelta perform delta and rate calculation
func (eo *EventOps) eventDelta(ctx *Context, event []byte) (out []byte, err error) {

	// Get any specified event unique attribute key, empty string otherwise
	keyVal, _ := js.GetUnsafeString(event, eo.Delta.KeyField)

	// Set the timestamp on the current event for caching
	if !js.Has(event, keys.Time) {
		if event, err = js.Set(event, ctx.CurrentRunTime(), keys.Time); err != nil {
			ctx.LogError("error setting current time %s", err.Error())
			return nil, err
		}
	}

	// Get previous event for delta.
	// If we can't find a existing event, store the current event and return
	key := append(deltaPrefix, []byte(ctx.name+"/"+ctx.node.HostName+"/"+keyVal)...)
	previous, err := ctx.txn.Get(key)

	if err != nil {
		if err == storage.ErrKeyNotFound {
			// ctx.LogDebug("event for key %s not found", key)
			// ctx.LogDebug("set event for key %s: %s", key, string(event))
			return nil, ctx.txn.SetWithTTL(key, event, eo.Delta.TTL)
		}
		return nil, err
	}

	// Store the current event
	if err = ctx.txn.SetWithTTL(key, event, eo.Delta.TTL); err != nil {
		return nil, err
	}

	// Parse the current and previous events, and timestamps for calculations
	previousTimestamp, err := js.GetTime(previous, keys.Time)
	if err != nil {
		return nil, err
	}
	timeDelta := ctx.CurrentRunTime().Sub(previousTimestamp).Seconds()

	// Loop over the event fields and perform the operations specified for each one.
	// if we get an error calculating stop return an error event to the stream
	err = js.ForEach(event, func(key string, value []byte) error {
		if key == keys.Time {
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
		previousValue, err := js.GetFloat(previous, key)
		if err != nil {
			return err
		}

		currentValue, err := js.GetFloat(value)
		if err != nil {
			return err
		}

		// ctx.LogDebug("delta for key: %s, current value: %x, previous value: %x",
		// key, currentValue, previousValue)

		// Get the delta for this key
		newValue := currentValue - previousValue

		// Check if have any rate calculation to do and if current field isn't blacklisted
		// If not calculate and round to the precision on FieldOps.Round
		if !eo.Delta.Rate {
			event, err = js.Set(event, newValue, key)
			return err
		}

		if _, ok := eo.Delta.RateBlacklist[key]; ok {
			event, err = js.Set(event, newValue, key)
			return err
		}

		event, err = js.Set(event, round(newValue/float64(timeDelta), eo.Round), key)
		return err
	})

	return event, err
}

// Round a float to the specified precision
func round(f float64, round int) (n float64) {
	shift := math.Pow(10, float64(round))
	return math.Floor((f*shift)+.5) / shift
}
