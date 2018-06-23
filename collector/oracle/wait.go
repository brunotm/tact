package oracle

import (
	"time"

	"github.com/brunotm/rexon"
	"github.com/brunotm/tact"
	"github.com/brunotm/tact/collector/client/oracle"
)

const (
	waitClassQuery = `select wait_class, total_waits as total_waits_avg,
	time_waited/10 as time_waited_ms_avg from v$system_wait_class where wait_class != 'Idle'`
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
		FieldTypes: []*rexon.Value{
			rexon.MustNewValue("wait_class", rexon.String),
			rexon.MustNewValue("total_waits_avg", rexon.Number),
			rexon.MustNewValue("time_waited_ms_avg", rexon.Number),
		},
		Delta: &tact.DeltaOps{
			Rate:      true,
			TTL:       15 * time.Minute,
			KeyField:  "wait_class",
			Blacklist: tact.BuildBlackList("wait_class"),
		},
	},
}

// vmstat collector
func waitClassFn(ctx *tact.Context) (events <-chan []byte) {
	return oracle.SingleQuery(ctx, waitClassQuery)
}
