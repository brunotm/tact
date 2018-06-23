package oracle

import (
	"github.com/brunotm/rexon"
	"github.com/brunotm/tact"
	"github.com/brunotm/tact/collector/client/oracle"
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
		FieldTypes: []*rexon.Value{
			rexon.MustNewValue("file_name", rexon.String),
			rexon.MustNewValue("file_id", rexon.Number),
			rexon.MustNewValue("tablespace_name", rexon.String),
			rexon.MustNewValue("bytes", rexon.DigitalUnit, rexon.ToFormat("mb")),
			rexon.MustNewValue("blocks", rexon.Number),
			rexon.MustNewValue("status", rexon.String),
			rexon.MustNewValue("relative_fno", rexon.Number),
			rexon.MustNewValue("autoextensible", rexon.String),
			rexon.MustNewValue("maxbytes", rexon.DigitalUnit, rexon.ToFormat("mb")),
			rexon.MustNewValue("maxblocks", rexon.Number),
			rexon.MustNewValue("increment_by", rexon.Number),
			rexon.MustNewValue("user_bytes", rexon.DigitalUnit, rexon.ToFormat("mb")),
			rexon.MustNewValue("user_blocks", rexon.Number),
			rexon.MustNewValue("online_status", rexon.String),
		},
	},
}

func datafilesFn(ctx *tact.Context) (events <-chan []byte) {
	return oracle.SingleQuery(ctx, datafilesQuery)
}
