package aix

import (
	"github.com/brunotm/tact"
	"github.com/brunotm/tact/collector/common"
)

const (
	inqPath = "/usr/lpp/EMC/Symmetrix/bin/inq.aix64_51"
	inqRex  = `^/dev/r(hdisk\d+)\s+(\w+)\s+(\w+)\s+(\w+)`
)

func init() {
	tact.Registry.Add(storage)
}

var storage = &tact.Collector{
	Name:    "/aix/config/storage",
	GetData: common.NewUnixEMCStorageFn(inqPath, inqRex),
}
