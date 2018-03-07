package linux

import (
	"github.com/brunotm/tact"
	"github.com/brunotm/tact/collector/common"
)

func init() {
	tact.Registry.Add(storage)
}

const (
	inqPath = "inq.LinuxAMD64"
	inqRex  = `^/dev/(.*?)\s+(\w+)\s+(\w+)\s+(\w+)`
)

var storage = &tact.Collector{
	Name:    "/linux/config/storage",
	GetData: common.NewUnixEMCStorageFn(inqPath, inqRex),
}
