package protocol

import (
	"testing"
	"time"

	"google.golang.org/protobuf/proto"
)

func TestMarshal(t *testing.T) {
	old := &Message{
		Header: &MessageHeader{
			ClientVersion: "go-libp2p/79da72fb7",
			Timestamp:     time.Now().Unix(),
			Id:            "abc123ABC",
			NodeId:        "qwer",
		},
		Type: *MesasgeType_PEER_IDENTITY_REQUEST.Enum(),
		Body: &Message_PiReq{
			PiReq: &PeerIdentityRequest{
				NodeId: "zxcv",
			},
		},
	}
	reqBytes, err := proto.Marshal(old)
	if err != nil {
		t.Fatalf("Marshal message failed %v", err)
	}
	new := &Message{}
	err = proto.Unmarshal(reqBytes, new)
	if err != nil {
		t.Fatalf("Unmarshal message failed %v", err)
	}
}
