package aix

import (
	"time"

	"fmt"

	"github.com/brunotm/rexon"
	"github.com/brunotm/tact"
	"github.com/brunotm/tact/collector"
)

const (
	errptCmd   = "/usr/bin/errpt -s "
	timeLayout = "0102150406"
)

// init register this collector with the dispatcher
func init() {
	tact.Registry.Add(errorLog)
}

var errorLog = &tact.Collector{
	Name:    "/aix/log/error",
	GetData: errorLogFn,
	PostOps: errorLogPostOps,
}

var errorLogParser = &rexon.RexLine{
	Rex:    rexon.RexMustCompile(`^(\w+)\s+(\d+)\s+(\w)\s+(\w)\s+(\w+)\s+(.*)`),
	Fields: []string{"identifier", tact.KeyTimeStamp, "type", "class", "resource", "description"},
	Types:  map[string]rexon.ValueType{rexon.KeyTypeAll: rexon.TypeString},
}

func errorLogFn(session tact.Session) <-chan []byte {

	// If this is our first run set back the clock to gather events
	timeLast := session.LastTime()
	if timeLast == 0 {
		timeLast = time.Now().AddDate(0, 0, -1).Unix()
	}

	return collector.SSHRex(session, errptCmd+time.Unix(timeLast, 0).Format(timeLayout), errorLogParser)
}

func errorLogPostOps(event []byte) ([]byte, error) {
	ts, _ := rexon.JSONGetUnsafeString(event, tact.KeyTimeStamp)
	timestamp, err := time.Parse(timeLayout, ts)
	if err != nil {
		return nil, fmt.Errorf("failed to parse timestamp: %s, error: %s", ts, err)
	}

	event, err = rexon.JSONSet(event, timestamp.Unix(), tact.KeyTimeStamp)
	return event, err
}
