package tact

import (
	"time"
)

// GetDataFn collect function type
type GetDataFn func(ctx *Context) (events <-chan []byte)

// PostEventOpsFn to post process events
type PostEventOpsFn func([]byte) (out []byte, err error)

// Collector implements the base collector and main routines
type Collector struct {
	Name     string
	GetData  GetDataFn
	EventOps *EventOps
	Joins    []*Join
	PostOps  PostEventOpsFn
}

// Start this collector with given ctxion and write channel
func (c *Collector) Start(ctx *Context, writeCh chan<- []byte) {
	defer ctx.cancel()

	// Build cache if needed
	err := c.buildRunCache(ctx)
	if err != nil {
		ctx.LogError("building cache", "error", err)
		return
	}

	events := c.GetData(ctx)

	for {
		select {
		case <-ctx.ctx.Done():
			deadline, ok := ctx.ctx.Deadline()
			if ok && deadline.Before(time.Now()) {
				ctx.LogWarn("context cancelled",
					"deadline", deadline.Format(time.RFC3339),
					"timeout_seconds", ctx.timeout)
			} else {
				ctx.LogWarn("context cancelled")
			}
			return

		case event, running := <-events:

			if !running {
				ctx.done()
				ctx.LogInfo("finished successfully")
				return
			}

			// Log if we received a null event from the collector and continue
			if event == nil {
				ctx.LogWarn("received null event")
				continue
			}

			// When performing delta ops the first time the returned event can be nil,
			// check to avoid getting empty events in the writer channel
			if c.EventOps != nil {
				if event = c.EventOps.process(ctx, event); event == nil {
					continue
				}
			}

			// Do any specified postOps
			if c.PostOps != nil {
				newEvent, err := c.PostOps(event)
				if err != nil {
					ctx.LogError("post ops error: %s, event: %s", err.Error(), string(event))
					continue
				}
				event = newEvent
			}

			// Do any specified data joins
			for _, join := range c.Joins {
				event, _ = join.Process(ctx, event)
			}

			// Enrich event with metadata from config and deliver to ctxion wchan
			event = ctx.enrichEvent(event)
			if !WrapCtxSend(ctx.ctx, writeCh, event) {
				ctx.LogError("timeout sending event to writer")
			}
		}
	}
}

func (c *Collector) buildRunCache(ctx *Context) (err error) {
	if len(c.Joins) > 0 {
		for _, join := range c.Joins {
			ctx.LogDebug("building join cache", "name", join.Name)
			if _, ok := ctx.cache[join.Name]; !ok {
				err := join.loadData(ctx)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}
