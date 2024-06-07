package types

type IdentifyProtocol struct {
	ID              string   `json:"peer_id"`
	ProtocolVersion string   `json:"protocol_version"`
	AgentVersion    string   `json:"agent_version"`
	Addresses       []string `json:"addresses"`
	Protocols       []string `json:"protocols"`
}
