package protocol

import (
	"testing"
	"time"

	"google.golang.org/protobuf/proto"
)

func TestMarshal(t *testing.T) {
	pi := &PeerIdentityBody{
		Data: &PeerIdentityBody_Req{
			Req: &PeerIdentityRequest{},
		},
	}
	body, err := proto.Marshal(pi)
	if err != nil {
		t.Fatalf("Marshal message body failed %v", err)
	}
	old := &Message{
		Header: &MessageHeader{
			ClientVersion: "go-libp2p/79da72fb7",
			Timestamp:     time.Now().Unix(),
			Id:            "abc123ABC",
			NodeId:        "qwer",
		},
		Type: *MessageType_PEER_IDENTITY.Enum(),
		Body: body,
	}
	reqBytes, err := proto.Marshal(old)
	if err != nil {
		t.Fatalf("Marshal message failed %v", err)
	}
	new := &Message{}
	if err := proto.Unmarshal(reqBytes, new); err != nil {
		t.Fatalf("Unmarshal message failed %v", err)
	}
	piBody := &PeerIdentityBody{}
	if err := proto.Unmarshal(new.Body, piBody); err != nil {
		t.Fatalf("Unmarshal message body failed %v", err)
	}
	if piReq := piBody.GetReq(); piReq != nil {
		t.Logf("Unmarshal Peer Identity Request success")
	} else {
		t.Fatal("Unmarshal message oneof error")
	}

	pib := &PeerIdentityBody{}
	if err := proto.Unmarshal(nil, pib); err != nil {
		t.Log("Unmarshal nil body return nil")
	}
}
