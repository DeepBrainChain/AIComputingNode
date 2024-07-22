package types

type IdentifyProtocol struct {
	ID              string   `json:"peer_id"`
	ProtocolVersion string   `json:"protocol_version"`
	AgentVersion    string   `json:"agent_version"`
	Addresses       []string `json:"addresses"`
	Protocols       []string `json:"protocols"`
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
