package oracle

import (
	"github.com/brunotm/rexon"
	"github.com/brunotm/tact"
)

const (
	datafilesQuery = `select * from dba_data_files`
)

// init register this collector with the dispatcher
func init() {
	tact.Registry.Add(datafiles)
}

var datafiles = &tact.Collector{
	Name:    "/oracle/config/datafiles",
	GetData: datafilesFn,
	EventOps: &tact.EventOps{
		Round: 2,
		FieldTypes: map[string]rexon.ValueType{
			rexon.KeyTypeAll:  rexon.TypeNumber,
			"file_name":       rexon.TypeString,
			"tablespace_name": rexon.TypeString,
			"status":          rexon.TypeString,
			"autoextensible":  rexon.TypeString,
			"online_status":   rexon.TypeString,
		},
	},
}

func datafilesFn(session *tact.Session) (events <-chan []byte) {
	return singleQuery(session, datafilesQuery)
}
