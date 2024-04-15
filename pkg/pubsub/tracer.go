package ps

import (
	"AIComputingNode/pkg/log"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	pb "github.com/libp2p/go-libp2p-pubsub/pb"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
)

type Tracer struct{}

func (t *Tracer) Trace(evt *pb.TraceEvent) {
	// log.Logger.Infof("Trace event: %v\n", evt)
	switch evt.GetType() {
	case pb.TraceEvent_PUBLISH_MESSAGE:
		log.Logger.Debugf("Trace event: {Type: %s, PeerID: %s, PublishMessage{MessageID: %s, Topic: %s}}",
			evt.Type, transform(evt.PeerID), string(evt.PublishMessage.MessageID), evt.PublishMessage.GetTopic())
	case pb.TraceEvent_REJECT_MESSAGE:
		log.Logger.Warnf("Trace event: {Type: %s, PeerID: %s, RejectMessage{MessageID: %s, ReceivedFrom: %s, Reason: %s, Topic: %s}}",
			evt.Type, transform(evt.PeerID), string(evt.RejectMessage.MessageID), transform(evt.RejectMessage.ReceivedFrom),
			evt.RejectMessage.GetReason(), evt.RejectMessage.GetTopic())
	case pb.TraceEvent_DUPLICATE_MESSAGE:
		log.Logger.Debugf("Trace event: {Type: %s, PeerID: %s, DuplicateMessage{MessageID: %s, ReceivedFrom: %s, Topic: %s}}",
			evt.Type, transform(evt.PeerID), string(evt.DuplicateMessage.MessageID), transform(evt.DuplicateMessage.ReceivedFrom),
			evt.DuplicateMessage.GetTopic())
	case pb.TraceEvent_DELIVER_MESSAGE:
		log.Logger.Debugf("Trace event: {Type: %s, PeerID: %s, DeliverMessage{MessageID: %s, ReceivedFrom: %s, Topic: %s}}",
			evt.Type, transform(evt.PeerID), string(evt.DeliverMessage.MessageID), transform(evt.DeliverMessage.ReceivedFrom),
			evt.DeliverMessage.GetTopic())
	case pb.TraceEvent_ADD_PEER:
		log.Logger.Infof("Trace event: {Type: %s, PeerID: %s, AddPeer{PeerID: %s, Proto: %s}}",
			evt.Type, transform(evt.PeerID), transform(evt.AddPeer.PeerID), evt.AddPeer.GetProto())
	case pb.TraceEvent_REMOVE_PEER:
		log.Logger.Infof("Trace event: {Type: %s, PeerID: %s, RemovePeer{PeerID: %s}}",
			evt.Type, transform(evt.PeerID), transform(evt.RemovePeer.PeerID))
	case pb.TraceEvent_RECV_RPC:
		log.Logger.Debugf("Trace event: {Type: %s, PeerID: %s, RecvRPC{ReceivedFrom: %s, Meta: %v}}",
			evt.Type, transform(evt.PeerID), transform(evt.RecvRPC.ReceivedFrom), evt.RecvRPC.Meta)
	case pb.TraceEvent_SEND_RPC:
		log.Logger.Debugf("Trace event: {Type: %s, PeerID: %s, SendRPC{SendTo: %s, Meta: %v}}",
			evt.Type, transform(evt.PeerID), transform(evt.SendRPC.SendTo), evt.SendRPC.Meta)
	case pb.TraceEvent_DROP_RPC:
		log.Logger.Infof("Trace event: {Type: %s, PeerID: %s, DropRPC{SendTo: %s, Meta: %v}}",
			evt.Type, transform(evt.PeerID), transform(evt.DropRPC.SendTo), evt.DropRPC.Meta)
	case pb.TraceEvent_JOIN:
		log.Logger.Infof("Trace event: {Type: %s, PeerID: %s, Join{Topic: %s}}",
			evt.Type, transform(evt.PeerID), evt.Join.GetTopic())
	case pb.TraceEvent_LEAVE:
		log.Logger.Infof("Trace event: {Type: %s, PeerID: %s, Leave{Topic: %s}}",
			evt.Type, transform(evt.PeerID), evt.Leave.GetTopic())
	case pb.TraceEvent_GRAFT:
		log.Logger.Infof("Trace event: {Type: %s, PeerID: %s, Graft{PeerID: %s, Topic: %s}}",
			evt.Type, transform(evt.PeerID), transform(evt.Graft.PeerID), evt.Graft.GetTopic())
	case pb.TraceEvent_PRUNE:
		log.Logger.Infof("Trace event: {Type: %s, PeerID: %s, Prune{PeerID: %s, Topic: %s}}",
			evt.Type, transform(evt.PeerID), transform(evt.Prune.PeerID), evt.Prune.GetTopic())
	default:
		log.Logger.Warn("Unknowned pubsub event type ", evt.GetType())
	}
}

func transform(bytes []byte) string {
	if peer, err := peer.IDFromBytes(bytes); err != nil {
		return string(bytes)
	} else {
		return peer.String()
	}
}

type RawTracer struct{}

func (t *RawTracer) AddPeer(p peer.ID, proto protocol.ID) {
	log.Logger.Infof("RawTrace AddPeer(%v, %v)", p.String(), proto)
}

func (t *RawTracer) RemovePeer(p peer.ID) {
	log.Logger.Infof("RawTrace RemovePeer(%v)", p.String())
}

func (t *RawTracer) Join(topic string) {
	log.Logger.Infof("RawTrace Join(%v)", topic)
}

func (t *RawTracer) Leave(topic string) {
	log.Logger.Infof("RawTrace Leave(%v)", topic)
}

func (t *RawTracer) Graft(p peer.ID, topic string) {
	log.Logger.Infof("RawTrace Graft(%v, %v)", p.String(), topic)
}

func (t *RawTracer) Prune(p peer.ID, topic string) {
	log.Logger.Infof("RawTrace Prune(%v, %v)", p.String(), topic)
}

func (t *RawTracer) ValidateMessage(msg *pubsub.Message) {
	log.Logger.Infof("RawTrace ValidateMessage{from: %v, seqno: %v}", string(msg.GetFrom()), string(msg.GetSeqno()))
}

func (t *RawTracer) DeliverMessage(msg *pubsub.Message) {
	log.Logger.Infof("RawTrace DeliverMessage{from: %v, seqno: %v}", string(msg.GetFrom()), string(msg.GetSeqno()))
}

func (t *RawTracer) RejectMessage(msg *pubsub.Message, reason string) {
	log.Logger.Infof("RawTrace RejectMessage({from: %v, seqno: %v}, %v)", string(msg.GetFrom()), string(msg.GetSeqno()), reason)
}

func (t *RawTracer) DuplicateMessage(msg *pubsub.Message) {
	log.Logger.Infof("RawTrace DuplicateMessage{from: %v, seqno: %v}", string(msg.GetFrom()), string(msg.GetSeqno()))
}

func (t *RawTracer) ThrottlePeer(p peer.ID) {
	log.Logger.Infof("RawTrace ThrottlePeer(%v)", p.String())
}

func (t *RawTracer) RecvRPC(rpc *pubsub.RPC) {
	log.Logger.Infof("RawTrace RecvRPC({%v})", rpc)
}

func (t *RawTracer) SendRPC(rpc *pubsub.RPC, p peer.ID) {
	log.Logger.Infof("RawTrace SendRPC({%v}, %v)", rpc, p.String())
}

func (t *RawTracer) DropRPC(rpc *pubsub.RPC, p peer.ID) {
	log.Logger.Infof("RawTrace DropRPC({%v}, %v)", rpc, p.String())
}

func (t *RawTracer) UndeliverableMessage(msg *pubsub.Message) {
	log.Logger.Infof("RawTrace UndeliverableMessage{from: %v, seqno: %v}", string(msg.GetFrom()), string(msg.GetSeqno()))
}
