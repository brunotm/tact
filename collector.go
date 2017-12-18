package tact

import (
	"time"
)

// Collector implements the base collector and main routines
type Collector struct {
	Name     string
	GetData  GetDataFn
	EventOps *EventOps
	Joins    []*Join
	PostOps  PostEventOpsFn
}

// Start this collector with given session and write channel
func (c *Collector) Start(sess *session, writeCh chan<- []byte) {
	sess.Start()
	defer sess.Cancel()

	// Build cache if needed
	err := c.buildRunCache(sess)
	if err != nil {
		sess.LogErr(err.Error())
		return
	}

	events := c.GetData(sess)

	for {
		select {
		case <-sess.ctx.Done():
			deadline, ok := sess.ctx.Deadline()
			if ok && deadline.Before(time.Now()) {
				sess.LogWarn("context cancelled, deadline: %s, timeout: %d seconds",
					deadline.Format(time.RFC3339), sess.timeout)
			} else {
				sess.LogWarn("context cancelled")
			}
			return

		case event, running := <-events:

			if !running {
				sess.Done()
				sess.LogInfo("finished successfuly")
				return
			}

			// Log if we received a null event from the collector and continue
			if event == nil {
				sess.LogWarn("received null event")
				continue
			}

			// When performing delta ops the first time the returned event can be nil,
			// check to avoid getting empty events in the writer channel
			if c.EventOps != nil {
				if event = c.EventOps.process(sess, event); event == nil {
					continue
				}
			}

			// Do any specified postOps
			if c.PostOps != nil {
				newEvent, err := c.PostOps(event)
				if err != nil {
					sess.LogErr("post ops processing error: %s, event: %s", err.Error(), string(event))
					continue
				}
				event = newEvent
			}

			// Do any specified data joins
			for _, join := range c.Joins {
				event, _ = join.Process(sess, event)
			}

			// Enrich event with metadata from config and deliver to session wchan
			event = sess.enrichEvent(event)
			if !WrapCtxSend(sess.ctx, writeCh, event) {
				sess.LogErr("timeout sending event to writer")
			}
		}
	}
}

func (c *Collector) buildRunCache(session *session) (err error) {
	if len(c.Joins) > 0 {
		for _, join := range c.Joins {
			err := join.loadData(session)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
