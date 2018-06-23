package linux

import (
	"time"

	"github.com/brunotm/tact/collector/keys"

	"github.com/brunotm/rexon"
	"github.com/brunotm/tact"
	"github.com/brunotm/tact/collector/client/ssh"
	"github.com/brunotm/tact/js"
)

func init() {
	tact.Registry.Add(lsblk)
	tact.Registry.Add(pvs)
	tact.Registry.Add(asmDevices)
}

var lsblk = &tact.Collector{
	Name:    "/linux/config/lsblk",
	GetData: lsblkFn,
	Joins: []*tact.Join{
		{
			TTL:           3 * time.Hour,
			Name:          "/linux/config/storage",
			JoinFields:    []string{keys.Device, keys.DeviceDM},
			JoinOnFields:  []string{keys.Device},
			IncludeFields: []string{keys.ArrayID, keys.ArrayDevice, keys.DeviceWWN},
		},
	},
}

var lsblkParser = rexon.MustNewParser(
	[]*rexon.Value{
		rexon.MustNewValue(keys.Device, rexon.String),
		rexon.MustNewValue(keys.DeviceDM, rexon.String),
		rexon.MustNewValue(keys.MajMin, rexon.String),
		rexon.MustNewValue(keys.SizeMB, rexon.DigitalUnit, rexon.ToFormat("mb")),
		rexon.MustNewValue(keys.MountPoint, rexon.String),
	},
	rexon.LineRegex(`(.*?)\s+\(?(.*?)\)?\s+(\d+:\d+)\s+\d+\s+(\d+)\s+\d+(.*)`),
)

func lsblkFn(ctx *tact.Context) (events <-chan []byte) {
	return ssh.Regex(ctx, "lsblk -lb", lsblkParser)
}

var pvs = &tact.Collector{
	Name:    "/linux/config/pvs",
	GetData: pvsFn,
	Joins: []*tact.Join{
		{
			TTL:           3 * time.Hour,
			Name:          "/linux/config/lsblk",
			JoinFields:    []string{keys.Device},
			JoinOnFields:  []string{keys.Device},
			IncludeFields: []string{keys.ArrayID, keys.ArrayDevice, keys.DeviceWWN, keys.SizeMB, keys.Device, keys.DeviceDM},
		},
	},
}

var pvsParser = rexon.MustNewParser(
	[]*rexon.Value{
		rexon.MustNewValue(keys.Device, rexon.String),
		rexon.MustNewValue(keys.VGName, rexon.String),
		rexon.MustNewValue(keys.VGType, rexon.String),
	},
	rexon.LineRegex(`/.*?/(\w+)\s+(\w+)\s+(\w+)`),
)

func pvsFn(ctx *tact.Context) (events <-chan []byte) {
	return ssh.Regex(ctx, "pvs", pvsParser)
}

var asmDevices = &tact.Collector{
	Name:    "/linux/config/asm",
	GetData: asmDevicesFn,
	PostOps: asmDevicesPostOpsFn,
	Joins: []*tact.Join{
		{
			TTL:           3 * time.Hour,
			Name:          "/linux/config/lsblk",
			JoinFields:    []string{keys.MajMin},
			JoinOnFields:  []string{keys.MajMin},
			IncludeFields: []string{keys.ArrayID, keys.ArrayDevice, keys.DeviceWWN, keys.SizeMB, keys.Device, keys.DeviceDM},
		},
	},
}

var asmDevicesParser = rexon.MustNewParser(
	[]*rexon.Value{
		rexon.MustNewValue(keys.Maj, rexon.String),
		rexon.MustNewValue(keys.Min, rexon.String),
		rexon.MustNewValue("asm_device", rexon.String),
	},
	rexon.LineRegex(`.*?\s+.*?\s+.*?\s+.*?(\d+),\s+(\d+)\s+.*?\s+.*?\s+.*?\s+.*?(.*)`),
)

func asmDevicesFn(ctx *tact.Context) (events <-chan []byte) {
	return ssh.Regex(ctx, "ls -l /dev/oracleasm/disks", asmDevicesParser)
}

func asmDevicesPostOpsFn(event []byte) (out []byte, err error) {
	major, _ := js.GetString(event, keys.Maj)
	minor, _ := js.GetString(event, keys.Min)
	event = js.Delete(event, keys.Maj)
	event = js.Delete(event, keys.Min)
	event, _ = js.Set(event, "oracleasm", keys.VGType)
	event, _ = js.Set(event, nil, keys.VGName)
	event, _ = js.Set(event, nil, keys.VGMode)
	return js.Set(event, major+":"+minor, keys.MajMin)
}
