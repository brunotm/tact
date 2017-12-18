package tact

import "context"

// Session carries the configuration and logging context
type Session interface {
	Name() string                                 // Return this session name
	Node() *Node                                  // Return this session node config
	Context() context.Context                     // Return this session context
	LastTime() int64                              // Return last succesful run timestamp
	LogInfo(message string, args ...interface{})  // LogInfo
	LogWarn(message string, args ...interface{})  // LogWarn
	LogErr(message string, args ...interface{})   // LogErr
	LogDebug(message string, args ...interface{}) // LogDebug

}

// Node holds configuration for the given session
type Node struct {
	HostName    string            `json:"hostname,omitempty"`
	NetAddr     string            `json:"netaddr,omitempty"`
	Type        string            `json:"type,omitempty"`
	SSHPort     string            `json:"ssh_port,omitempty"`
	SSHUser     string            `json:"ssh_user,omitempty"`
	SSHPassword string            `json:"ssh_password,omitempty"`
	SSHKey      []byte            `json:"ssh_key,omitempty"`
	APIURL      string            `json:"api_url,omitempty"`
	APIUser     string            `json:"api_user,omitempty"`
	APIPassword string            `json:"api_password,omitempty"`
	DBUser      string            `json:"db_user,omitempty"`
	DBPassword  string            `json:"db_password,omitempty"`
	DBPort      string            `json:"db_port,omitempty"`
	LogFiles    map[string]string `json:"files,omitempty"`
}

// GetDataFn collect function type
type GetDataFn func(session Session) <-chan []byte

// PostEventOpsFn to post process events
type PostEventOpsFn func([]byte) ([]byte, error)
