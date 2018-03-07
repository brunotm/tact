package oracle

import (
	"github.com/brunotm/rexon"
	"github.com/brunotm/tact"
)

const (
	waitClassQuery = `select wait_class, total_waits, time_waited/100 as time_waited from v$system_wait_class where wait_class != 'Idle'`
)

// init register this collector with the dispatcher
func init() {
	tact.Registry.Add(waitClass)
}

var waitClass = &tact.Collector{
	Name:    "/oracle/performance/waitclass",
	GetData: waitClassFn,
	EventOps: &tact.EventOps{
		Round: 2,
		FieldTypes: map[string]rexon.ValueType{
			rexon.KeyTypeAll: rexon.TypeNumber,
			"wait_class":     rexon.TypeString,
		},
		Delta: &tact.DeltaOps{
			Rate:      true,
			KeyField:  "wait_class",
			Blacklist: tact.BuildBlackList("wait_class"),
		},
	},
}

// vmstat collector
func waitClassFn(session *tact.Session) (events <-chan []byte) {
	return singleQuery(session, waitClassQuery)
}
