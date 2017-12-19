package linux

import (
	"github.com/brunotm/tact"
	"github.com/brunotm/tact/collector/common"
)

// init register this collector with the dispatcher
func init() {
	tact.Registry.Add(&tact.Collector{
		Name:    "/linux/performance/uptime",
		GetData: common.NewUnixUptimeFn(),
	})
}
