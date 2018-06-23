package oracle

import (
	"github.com/brunotm/rexon"
	"github.com/brunotm/tact"
	"github.com/brunotm/tact/collector/client/oracle"
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
		FieldTypes: []*rexon.Value{
			rexon.MustNewValue("tablespace_name", rexon.String),
			rexon.MustNewValue("block_size", rexon.Number),
			rexon.MustNewValue("initial_extent", rexon.Number),
			rexon.MustNewValue("next_extent", rexon.Number, rexon.Nullable()),
			rexon.MustNewValue("min_extents", rexon.Number),
			rexon.MustNewValue("max_extents", rexon.Number, rexon.Nullable()),
			rexon.MustNewValue("max_size", rexon.Number),
			rexon.MustNewValue("pct_increase", rexon.Number, rexon.Nullable()),
			rexon.MustNewValue("min_extlen", rexon.Number),
			rexon.MustNewValue("status", rexon.String),
			rexon.MustNewValue("contents", rexon.String),
			rexon.MustNewValue("logging", rexon.String),
			rexon.MustNewValue("force_logging", rexon.String),
			rexon.MustNewValue("extent_management", rexon.String),
			rexon.MustNewValue("allocation_type", rexon.String),
			rexon.MustNewValue("plugged_in", rexon.String),
			rexon.MustNewValue("segment_space_management", rexon.String),
			rexon.MustNewValue("def_tab_compression", rexon.String),
			rexon.MustNewValue("retention", rexon.String),
			rexon.MustNewValue("bigfile", rexon.String),
			rexon.MustNewValue("predicate_evaluation", rexon.String),
			rexon.MustNewValue("encrypted", rexon.String),
			rexon.MustNewValue("compress_for", rexon.String, rexon.Nullable()),
		},
	},
}

func tableSpacesFn(ctx *tact.Context) (events <-chan []byte) {
	return oracle.SingleQuery(ctx, tableSpacesQuery)
}
