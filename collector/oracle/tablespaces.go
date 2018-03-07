package oracle

import (
	"github.com/brunotm/rexon"
	"github.com/brunotm/tact"
)

const (
	tableSpacesQuery = `select * from dba_tablespaces`
)

func init() {
	tact.Registry.Add(tableSpaces)
}

var tableSpaces = &tact.Collector{
	Name:    "/oracle/config/tablespaces",
	GetData: tableSpacesFn,
	EventOps: &tact.EventOps{
		Round: 2,
		FieldTypes: map[string]rexon.ValueType{
			rexon.KeyTypeAll: rexon.TypeString,
			"block_size":     rexon.TypeNumber,
			"initial_extent": rexon.TypeNumber,
			"next_extent":    rexon.TypeNumber,
			"min_extents":    rexon.TypeNumber,
			"max_extents":    rexon.TypeNumber,
			"max_size":       rexon.TypeNumber,
			"pct_increase":   rexon.TypeNumber,
			"min_extlen":     rexon.TypeNumber,
		},
	},
}

func tableSpacesFn(session *tact.Session) (events <-chan []byte) {
	return singleQuery(session, tableSpacesQuery)
}
