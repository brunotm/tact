package aix

import (
	"github.com/brunotm/tact"
	"github.com/brunotm/tact/collector/common"
)

// init register this collector with the dispatcher
func init() {
	tact.Registry.Add(upTime)
}

var upTime = &tact.Collector{
	Name:    "/aix/performance/uptime",
	GetData: common.NewUnixUptimeFn(),
}
