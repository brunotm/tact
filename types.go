package tact

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
