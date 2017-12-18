package collector

import (
	"bufio"
	"bytes"
	"io"

	"github.com/brunotm/rexon"
	"github.com/brunotm/sshmgr"
	"github.com/brunotm/tact"
)

const (
	buffered = false
)

// SSHScanner collector base
func SSHScanner(session tact.Session, cmd string) *bufio.Scanner {
	sshSession, err := sshmgr.Manager.GetSSHSession(NewSSHNodeConfig(session.Node()))
	if err != nil {
		session.LogErr("sshscanner: error getting ssh session: %s", err.Error())
		return nil
	}
	defer sshSession.Close()

	if buffered {
		data, err := sshSession.CombinedOutput(cmd)
		if err != nil {
			session.LogErr("executing ssh comand: %s, error: %s", cmd, err.Error())
			return nil
		}
		return bufio.NewScanner(bytes.NewReader(data))
	}

	stdout, err := sshSession.StdoutPipe()
	if err != nil {
		session.LogErr("sshscanner: getting stdout pipe: %s", err.Error())
		return nil
	}
	stderr, err := sshSession.StderrPipe()
	if err != nil {
		session.LogErr("sshscanner: getting stderr pipe: %s", err.Error())
		return nil
	}

	// Combine both stdout and stderr
	// Send the data channel to the parser
	data := io.MultiReader(stdout, stderr)

	if err := sshSession.Run(cmd); err != nil {
		session.LogErr("sshrex: executing command: %s, error", cmd, err.Error())
		return nil
	}

	return bufio.NewScanner(data)
}

// SSHRex collector base
func SSHRex(session tact.Session, cmd string, rex rexon.Parser) <-chan []byte {
	outChan := make(chan []byte)
	go sshRex(session, cmd, rex, outChan)
	return outChan
}

func sshRex(session tact.Session, cmd string, rex rexon.Parser, outChan chan<- []byte) {
	defer close(outChan)

	sshSession, err := sshmgr.Manager.GetSSHSession(NewSSHNodeConfig(session.Node()))
	if err != nil {
		session.LogErr("sshrex: error getting ssh session: %s", err.Error())
		return
	}
	defer sshSession.Close()

	var rchan <-chan rexon.Result

	// If working in buffered mode, fetch all data into a []byte before feeding it to the parser.
	// Else working the data stream combining both StdoutPipe and StderrPipe into a io.MultiReader
	if buffered {
		data, err := sshSession.CombinedOutput(cmd)
		if err != nil {
			session.LogErr("sshrex: executing command: %s, error", cmd, err.Error())
			return
		}
		// Send the data channel to the parser
		rchan = rex.ParseBytes(session.Context(), data)
	} else {
		stdout, err := sshSession.StdoutPipe()
		if err != nil {
			session.LogErr("sshrex: getting stdout pipe: %s", err.Error())
			return
		}
		stderr, err := sshSession.StderrPipe()
		if err != nil {
			session.LogErr("sshrex: getting stderr pipe: %s", err.Error())
			return
		}
		// Combine both stdout and stderr
		// Send the data channel to the parser
		data := io.MultiReader(stdout, stderr)
		rchan = rex.Parse(session.Context(), data)

		if err := sshSession.Run(cmd); err != nil {
			session.LogErr("sshrex: executing command: %s, error", cmd, err.Error())
			return
		}
	}

	// Fetch and send parsed events upstream
	for result := range rchan {

		for e := range result.Errors {
			session.LogErr(result.Errors[e].Error())
		}

		if result.Data == nil {
			continue
		}

		if !tact.WrapCtxSend(session.Context(), outChan, result.Data) {
			session.LogErr("sshrex: timed out sending event to upstream processing")
			return
		}
	}
}

// NewSSHNodeConfig creates a SSHConfig from a NodeConfig
func NewSSHNodeConfig(node *tact.Node) *sshmgr.SSHConfig {
	return sshmgr.NewConfig(node.NetAddr, node.SSHUser, node.SSHPassword, node.SSHKey)
}
