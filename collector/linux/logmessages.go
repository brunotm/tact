package linux

import (
	"fmt"
	"time"

	"github.com/brunotm/rexon"
	"github.com/brunotm/tact"
	"github.com/brunotm/tact/collector"
)

const (
	fileName   = "messages"
	timeLayout = "Jan 2 15:04:05 2006"
)

// add collector to registry
func init() {
	tact.Registry.Add(logMessages)
}

var logMessages = &tact.Collector{
	Name:    "/linux/log/messages",
	GetData: logMessagesFn,
	PostOps: logMessagesPostOps,
}

var logMessagesParser = &rexon.RexLine{
	Prep:   rexon.RexMustCompile(`[(),;!"']`),
	Rex:    rexon.RexMustCompile(`^(\w{3}\s+\d+\s+\d{2}:\d{2}:\d{2})\s+\w+\s+(.*)\[(\d+)\]:\s+(.*)`),
	Fields: []string{tact.KeyTimeStamp, "resource", "pid", "message"},
	Types: map[string]rexon.ValueType{
		rexon.KeyTypeAll: rexon.TypeString,
		"pid":            rexon.TypeNumber},
}

func logMessagesFn(session tact.Session) (outCh <-chan []byte) {
	return collector.SFTPRex(session, fileName, logMessagesParser)
}

func logMessagesPostOps(event []byte) ([]byte, error) {
	ts, _ := rexon.JSONGetUnsafeString(event, tact.KeyTimeStamp)
	timestamp, err := time.Parse(timeLayout, fmt.Sprintf("%s %d", ts, time.Now().Year()))
	if err != nil {
		return nil, fmt.Errorf("failed to parse timestamp: %s, error: %s", ts, err.Error())
	}
	event, err = rexon.JSONSet(event, timestamp.Unix(), tact.KeyTimeStamp)
	return event, err
}
