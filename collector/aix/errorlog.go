package aix

import (
	"time"

	"github.com/brunotm/tact/collector/keys"

	"github.com/brunotm/rexon"
	"github.com/brunotm/tact"
	"github.com/brunotm/tact/collector/client/ssh"
)

const (
	errptCmd   = "/usr/bin/errpt -s "
	timeLayout = "0102150406"
)

func init() {
	tact.Registry.Add(errorLog)
}

var errorLog = &tact.Collector{
	Name:    "/aix/log/error",
	GetData: errorLogFn,
}

// Example raw data
// IDENTIFIER TIMESTAMP  T C RESOURCE_NAME  DESCRIPTION
// E87EF1BE   0604150018 P O dumpcheck      The largest dump device is too small.
// A924A5FC   0604130118 P S SYSPROC        SOFTWARE PROGRAM ABNORMALLY TERMINATED
// A924A5FC   0603230218 P S SYSPROC        SOFTWARE PROGRAM ABNORMALLY TERMINATED
// A924A5FC   0603192518 P S SYSPROC        SOFTWARE PROGRAM ABNORMALLY TERMINATED
// A924A5FC   0603150118 P S SYSPROC        SOFTWARE PROGRAM ABNORMALLY TERMINATED

var errorLogParser = rexon.MustNewParser(
	[]*rexon.Value{
		rexon.MustNewValue("identifier", rexon.String),
		rexon.MustNewValue(keys.Time, rexon.Time, rexon.FromFormat(timeLayout)),
		rexon.MustNewValue("type", rexon.String),
		rexon.MustNewValue("class", rexon.String),
		rexon.MustNewValue("resource", rexon.String),
		rexon.MustNewValue("description", rexon.String),
	},
	rexon.LineRegex(`(\w+)\s+(\d+)\s+(\w)\s+(\w)\s+(\w+)\s+(.*)`),
)

func errorLogFn(ctx *tact.Context) (events <-chan []byte) {

	// If this is our first run set back the clock 1 day to gather events
	timeLast := ctx.LastRunTime()
	if timeLast.IsZero() {
		timeLast = time.Now().AddDate(0, 0, -1)
	}

	return ssh.Regex(ctx, errptCmd+timeLast.Format(timeLayout), errorLogParser)
}
