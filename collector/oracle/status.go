package oracle

import (
	"github.com/brunotm/rexon"
	"github.com/brunotm/tact"
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
		Round: 2,
		FieldTypes: map[string]rexon.ValueType{
			rexon.KeyTypeAll: rexon.TypeNumber,
		},
	},
}

// vmstat collector
func sessionsFn(session *tact.Session) (events <-chan []byte) {
	return singleQuery(session, sessionsQuery)
}
