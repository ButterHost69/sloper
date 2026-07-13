package agent

// rpcCmd is the JSON shape sent to pi --mode rpc on stdin.
type rpcCmd struct {
	ID      string `json:"id,omitempty"`
	Type    string `json:"type"`
	Message string `json:"message,omitempty"`
}

func newRPCCmd(typ, msg string) rpcCmd {
	return rpcCmd{
		Type:    typ,
		Message: msg,
	}
}

func (c rpcCmd) withID(id string) rpcCmd {
	c.ID = id
	return c
}
