package types

const ChatProxyProtocol = "/chat-proxy/0.0.1"

type IdentifyProtocol struct {
	ID              string   `json:"peer_id,omitempty"`
	ProtocolVersion string   `json:"protocol_version,omitempty"`
	AgentVersion    string   `json:"agent_version,omitempty"`
	Addresses       []string `json:"addresses,omitempty"`
	Protocols       []string `json:"protocols,omitempty"`
}

type NodeType uint32

const (
	PublicIpFlag     NodeType = 0x01
	PeersCollectFlag NodeType = 0x02
	ModelFlag        NodeType = 0x04
)

func (nt NodeType) IsPublicNode() bool {
	return nt&PublicIpFlag != 0
}

func (nt NodeType) IsClientNode() bool {
	return nt&PeersCollectFlag != 0
}

func (nt NodeType) IsModelNode() bool {
	return nt&ModelFlag != 0
}
