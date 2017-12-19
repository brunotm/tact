package oracle

import (
	"fmt"
	"time"

	"github.com/brunotm/rexon"
	"github.com/brunotm/tact"
	"github.com/brunotm/tact/collector"
)

const (
	fileName   = "alert"
	timeLayout = "2006-01-02T15:04:05.999-07:00"
)

// init register this collector with the dispatcher
func init() {
	tact.Registry.Add(logAlert)
}

var logAlert = &tact.Collector{
	Name:    "/oracle/log/alert",
	GetData: logAlertFn,
	PostOps: logAlertPostOps,
}

var logAlertParser = &rexon.RexSet{
	Round: 2,
	Types: map[string]rexon.ValueType{
		rexon.KeyTypeAll: rexon.TypeString,
		"pid":            rexon.TypeNumber},
	Set: rexon.RexSetMustCompile(map[string]string{
		rexon.KeyStartTag: `^<msg`,
		tact.KeyTimeStamp: `time='(.*?)'`,
		"org_id":          `org_id='(.*?)'`,
		"comp_id":         `comp_id='(.*?)'`,
		"client_id":       `client_id='(.*?)'`,
		"type":            `type='(.*?)'`,
		"level":           `level='(.*?)'`,
		"host_id":         `host_id='(.*?)'`,
		"host_addr":       `host_addr='(.*?)'`,
		"module":          `module='(.*?)'`,
		"pid":             `pid='(.*?)'`,
		"message":         `<txt>\s*(.*)`,
	}),
}

func logAlertFn(session tact.Session) <-chan []byte {
	return collector.SFTPRex(session, fileName, logAlertParser)
}

func logAlertPostOps(event []byte) ([]byte, error) {
	ts, _ := rexon.JSONGetUnsafeString(event, tact.KeyTimeStamp)
	timestamp, err := time.Parse(timeLayout, ts)
	if err != nil {
		return nil, fmt.Errorf("failed to parse timestamp: %s, error: %s", ts, err.Error())
	}
	event, err = rexon.JSONSet(event, timestamp.Unix(), tact.KeyTimeStamp)
	return event, err
}
