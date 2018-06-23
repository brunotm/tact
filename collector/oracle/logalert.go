package oracle

import (
	"github.com/brunotm/rexon"
	"github.com/brunotm/tact"
	"github.com/brunotm/tact/collector/client/sftp"
	"github.com/brunotm/tact/collector/keys"
)

const (
	fileName   = "alert"
	timeLayout = "2006-01-02T15:04:05.999-07:00"
)

// init register this collector with the dispatcher
func init() {
	tact.Registry.Add(logAlert)
}

var logAlert = &tact.Collector{
	Name:    "/oracle/log/alert",
	GetData: logAlertFn,
}

var logAlertParser = rexon.MustNewParser(
	[]*rexon.Value{
		rexon.MustNewValue(keys.Time, rexon.Time, rexon.FromFormat(timeLayout), rexon.ValueRegex(`time='(.*?)'`)),
		rexon.MustNewValue("org_id", rexon.String, rexon.ValueRegex(`org_id='(.*?)'`)),
		rexon.MustNewValue("comp_id", rexon.String, rexon.ValueRegex(`comp_id='(.*?)'`)),
		rexon.MustNewValue("client_id", rexon.String, rexon.ValueRegex(`client_id='(.*?)'`)),
		rexon.MustNewValue("type", rexon.String, rexon.ValueRegex(`type='(.*?)'`)),
		rexon.MustNewValue("level", rexon.String, rexon.ValueRegex(`level='(.*?)'`)),
		rexon.MustNewValue("host_id", rexon.String, rexon.ValueRegex(`host_id='(.*?)'`)),
		rexon.MustNewValue("host_addr", rexon.String, rexon.ValueRegex(`host_addr='(.*?)'`)),
		rexon.MustNewValue("module", rexon.String, rexon.ValueRegex(`module='(.*?)'`)),
		rexon.MustNewValue("pid", rexon.String, rexon.ValueRegex(`pid='(.*?)'`)),
		rexon.MustNewValue("message", rexon.String, rexon.ValueRegex(`<txt>\s*(.*)`)),
	},
	rexon.StartTag(`^<msg`),
)

func logAlertFn(ctx *tact.Context) (events <-chan []byte) {
	return sftp.Regex(ctx, fileName, logAlertParser)
}
