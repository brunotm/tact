package ssh

import (
	"time"

	"github.com/brunotm/rexon"
	"github.com/brunotm/sshmgr"
	"github.com/brunotm/sshmgr/manager"
	"github.com/brunotm/tact"
)

const (
	buffered = false
)

// Client creates a new ssh client
func Client(ctx *tact.Context) (client *sshmgr.Client, err error) {
	return manager.SSHClient(NewSSHNodeConfig(ctx))
}

// Regex executes the given command and parses with the provided parser
func Regex(ctx *tact.Context, cmd string, rex rexon.DataParser) (events <-chan []byte) {
	outCh := make(chan []byte)
	go regex(ctx, cmd, rex, outCh)
	return outCh
}

func regex(ctx *tact.Context, cmd string, rex rexon.DataParser, outCh chan<- []byte) {
	defer close(outCh)

	client, err := manager.SSHClient(NewSSHNodeConfig(ctx))
	if err != nil {
		ctx.LogError("sshrex: error getting ssh client: %s", err.Error())
		return
	}
	defer client.Close()

	var rchan <-chan rexon.Result

	// If working in buffered mode, fetch all data into a []byte before feeding it to the parser.
	// Else working the data stream combining both StdoutPipe and StderrPipe into a io.MultiReader
	if buffered {
		data, err := client.CombinedOutput(cmd, nil)
		if err != nil {
			ctx.LogError("sshrex: executing command: %s, error", cmd, err.Error())
			return
		}

		// Send the data channel to the parser
		rchan = rex.ParseBytes(ctx.Context(), data)

	} else {
		data, err := client.CombinedReader(cmd, nil)
		if err != nil {
			ctx.LogError("sshrex: executing command: %s, error", cmd, err.Error())
			return
		}
		defer data.Close()

		rchan = rex.Parse(ctx.Context(), data)
	}

	// Fetch and send parsed events upstream
	for result := range rchan {

		for e := range result.Errors {
			ctx.LogError(result.Errors[e].Error())
		}

		if result.Data == nil {
			continue
		}

		if !tact.WrapCtxSend(ctx.Context(), outCh, result.Data) {
			ctx.LogError("sshrex: timed out sending event to upstream processing")
			return
		}
	}
}

// NewSSHNodeConfig creates a SSHConfig from a NodeConfig
func NewSSHNodeConfig(ctx *tact.Context) (config sshmgr.ClientConfig) {
	node := ctx.Node()
	config.NetAddr = node.NetAddr
	config.Port = node.SSHPort
	config.User = node.SSHUser
	config.Password = node.SSHPassword
	config.Key = node.SSHKey
	config.IgnoreHostKey = true
	config.ConnDeadline = ctx.Timeout()
	config.DialTimeout = time.Second * 5
	return config
}
