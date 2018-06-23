package oracle

import (
	"github.com/brunotm/rexon"
	"github.com/brunotm/tact"
	"github.com/brunotm/tact/collector/client/oracle"
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
		FieldTypes: []*rexon.Value{
			rexon.MustNewValue("group_number", rexon.Number),
			rexon.MustNewValue("name", rexon.String),
			rexon.MustNewValue("sector_size", rexon.Number),
			rexon.MustNewValue("block_size", rexon.Number),
			rexon.MustNewValue("allocation_unit_size", rexon.Number),
			rexon.MustNewValue("state", rexon.String),
			rexon.MustNewValue("type", rexon.String),
			rexon.MustNewValue("total_mb", rexon.Number),
			rexon.MustNewValue("free_mb", rexon.Number),
			rexon.MustNewValue("hot_used_mb", rexon.Number),
			rexon.MustNewValue("cold_used_mb", rexon.Number),
			rexon.MustNewValue("required_mirror_free_mb", rexon.Number),
			rexon.MustNewValue("usable_file_mb", rexon.Number),
			rexon.MustNewValue("offline_disks", rexon.Number),
			rexon.MustNewValue("compatibility", rexon.String),
			rexon.MustNewValue("database_compatibility", rexon.String),
			rexon.MustNewValue("voting_files", rexon.String),
		},
	},
}

func asmDiskGroupFn(ctx *tact.Context) (events <-chan []byte) {
	return oracle.SingleQuery(ctx, asmDiskGroupQuery)
}
