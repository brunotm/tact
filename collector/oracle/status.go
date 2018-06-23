package oracle

import (
	"github.com/brunotm/rexon"
	"github.com/brunotm/tact"
	"github.com/brunotm/tact/collector/client/oracle"
)

const (
	sessionsQuery = `SELECT COUNT(*) as active_sessions FROM v$session WHERE status = 'INACTIVE'`
)

func init() {
	tact.Registry.Add(sessions)
}

var sessions = &tact.Collector{
	Name:    "/oracle/performance/sessions",
	GetData: sessionsFn,
	EventOps: &tact.EventOps{
		FieldTypes: []*rexon.Value{
			rexon.MustNewValue("active_sessions", rexon.Number),
		},
	},
}

// vmstat collector
func sessionsFn(ctx *tact.Context) (events <-chan []byte) {
	return oracle.SingleQuery(ctx, sessionsQuery)
}
