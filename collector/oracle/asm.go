package oracle

import (
	"github.com/brunotm/rexon"
	"github.com/brunotm/tact"
)

const (
	asmDiskGroupQuery = `select * from v$asm_diskgroup`
)

func init() {
	tact.Registry.Add(asmDiskGroup)
}

var asmDiskGroup = &tact.Collector{
	Name:    "/oracle/config/asm_diskgroup",
	GetData: asmDiskGroupFn,
	EventOps: &tact.EventOps{
		Round: 2,
		FieldTypes: map[string]rexon.ValueType{
			rexon.KeyTypeAll:         rexon.TypeNumber,
			"name":                   rexon.TypeString,
			"state":                  rexon.TypeString,
			"type":                   rexon.TypeString,
			"status":                 rexon.TypeString,
			"compatibility":          rexon.TypeString,
			"database_compatibility": rexon.TypeString,
			"voting_files":           rexon.TypeString,
		},
	},
}

func asmDiskGroupFn(session *tact.Session) (events <-chan []byte) {
	return singleQuery(session, asmDiskGroupQuery)
}
