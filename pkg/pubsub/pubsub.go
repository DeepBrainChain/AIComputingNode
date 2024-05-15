package ps

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"AIComputingNode/pkg/config"
	"AIComputingNode/pkg/db"
	"AIComputingNode/pkg/host"
	"AIComputingNode/pkg/ipfs"
	"AIComputingNode/pkg/log"
	"AIComputingNode/pkg/p2p"
	"AIComputingNode/pkg/protocol"
	"AIComputingNode/pkg/serve"
	"AIComputingNode/pkg/types"

	"google.golang.org/protobuf/proto"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
)

var MsgNotSupported string = "Not supported"

type ModelRequest struct {
	Prompt string `json:"prompt"`
}

type ModelResponse struct {
	Code     int    `json:"code"`
	Status   string `json:"status"`
	ImageUrl string `json:"imageUrl"`
}

func PublishToTopic(ctx context.Context, topic *pubsub.Topic, messageChan <-chan []byte) {
	for message := range messageChan {
		if err := topic.Publish(ctx, message); err != nil {
			log.Logger.Errorf("Error when publish to topic %v", err)
		} else {
			log.Logger.Infof("Published %d bytes", len(message))
		}
	}
}

func PubsubHandler(ctx context.Context, sub *pubsub.Subscription, publishChan chan<- []byte) {
	defer sub.Cancel()
	for {
		msg, err := sub.Next(ctx)
		if err != nil {
			log.Logger.Warnf("Read PubSub: %v", err)
			continue
		}

		pmsg := &protocol.Message{}
		if err := proto.Unmarshal(msg.Data, pmsg); err != nil {
			log.Logger.Warnf("Unmarshal PubSub: %v", err)
			continue
		}

		if pmsg.Header.NodeId == config.GC.Identity.PeerID {
			log.Logger.Infof("Received message type %s from the node itself", pmsg.Type)
			continue
		} else if pmsg.Header.Receiver != config.GC.Identity.PeerID {
			log.Logger.Infof("Gossip message type %s from %s to %s", pmsg.Type, pmsg.Header.NodeId, pmsg.Header.Receiver)
			continue
		} else {
			log.Logger.Infof("Received message type %s from %s", pmsg.Type, pmsg.Header.NodeId)
		}

		msgBody, err := p2p.Decrypt(pmsg.Header.NodePubKey, pmsg.Body)
		if err != nil {
			log.Logger.Warnf("Decrypt message body failed %v", err)
			continue
		}

		switch pmsg.Type {
		case protocol.MessageType_PEER_IDENTITY:
			pi := &protocol.PeerIdentityBody{}
			if pmsg.ResultCode != 0 {
				res := serve.PeerResponse{
					Code:    int(pmsg.ResultCode),
					Message: pmsg.ResultMessage,
					Data:    p2p.IdentifyProtocol{},
				}
				notifyData, err := json.Marshal(res)
				if err != nil {
					log.Logger.Warnf("Marshal Identity Protocol %v", err)
					break
				}
				serve.WriteAndDeleteRequestItem(pmsg.Header.Id, notifyData)
			} else if err := proto.Unmarshal(msgBody, pi); err == nil {
				if piReq := pi.GetReq(); piReq != nil {
					if piReq.GetNodeId() == config.GC.Identity.PeerID {
						idp := p2p.Hio.GetIdentifyProtocol()
						piBody := &protocol.PeerIdentityBody{
							Data: &protocol.PeerIdentityBody_Res{
								Res: &protocol.PeerIdentityResponse{
									ProtocolVersion: idp.ProtocolVersion,
									AgentVersion:    idp.AgentVersion,
									ListenAddrs:     idp.Addresses,
									Protocols:       idp.Protocols,
								},
							},
						}
						resBody, err := proto.Marshal(piBody)
						if err != nil {
							log.Logger.Warnf("Marshal Identity Response Body %v", err)
							break
						}
						resBody, err = p2p.Encrypt(pmsg.Header.NodeId, resBody)
						res := protocol.Message{
							Header: &protocol.MessageHeader{
								ClientVersion: p2p.Hio.UserAgent,
								Timestamp:     time.Now().Unix(),
								Id:            pmsg.Header.Id,
								NodeId:        config.GC.Identity.PeerID,
								Receiver:      pmsg.Header.NodeId,
							},
							Type:       protocol.MessageType_PEER_IDENTITY,
							Body:       resBody,
							ResultCode: 0,
						}
						if err == nil {
							res.Header.NodePubKey, _ = p2p.MarshalPubKeyFromPrivKey(p2p.Hio.PrivKey)
						}
						resBytes, err := proto.Marshal(&res)
						if err != nil {
							log.Logger.Warnf("Marshal Identity Response %v", err)
							break
						}
						publishChan <- resBytes
						log.Logger.Info("Sending Peer Identity Response")
					} else {
						log.Logger.Info("Gossip Peer Identity Request of ", piReq.GetNodeId())
					}
				} else if piRes := pi.GetRes(); piRes != nil {
					res := serve.PeerResponse{
						Code:    0,
						Message: "ok",
						Data: p2p.IdentifyProtocol{
							ID:              pmsg.Header.NodeId,
							ProtocolVersion: piRes.ProtocolVersion,
							AgentVersion:    piRes.AgentVersion,
							Addresses:       piRes.ListenAddrs,
							Protocols:       piRes.Protocols,
						},
					}
					notifyData, err := json.Marshal(res)
					if err != nil {
						log.Logger.Warnf("Marshal Identity Protocol %v", err)
						break
					}
					serve.WriteAndDeleteRequestItem(pmsg.Header.Id, notifyData)
				}
			} else {
				log.Logger.Warn("Message type and body do not match")
			}
		case protocol.MessageType_IMAGE_GENERATION:
			ig := &protocol.ImageGenerationBody{}
			if pmsg.ResultCode != 0 {
				res := serve.ImageGenerationResponse{
					Code:    int(pmsg.ResultCode),
					Message: pmsg.ResultMessage,
				}
				notifyData, err := json.Marshal(res)
				if err != nil {
					log.Logger.Warnf("Marshal Identity Protocol %v", err)
					break
				}
				serve.WriteAndDeleteRequestItem(pmsg.Header.Id, notifyData)
			} else if err := proto.Unmarshal(msgBody, ig); err == nil {
				if igReq := ig.GetReq(); igReq != nil {
					if igReq.GetNodeId() == config.GC.Identity.PeerID {
						// TODO: Run the model and upload the images generated by the model to the IPFS node
						// var ipfsAddr string = "/ip4/192.168.1.159/tcp/4002"
						var (
							code           = 0
							msg            = "ok"
							filePath       = ""
							cid            = ""
							err      error = nil
						)
						var ipfsAddr string = igReq.GetIpfsNode()
						if ipfsAddr == "" {
							ipfsAddr = config.GC.App.IpfsStorageAPI
						}
						if ipfsAddr == "" {
							code = int(types.ErrCodeModel)
							msg = "No ipfs storage server"
						} else {
							if config.ValidateIpfsServer(ipfsAddr) {
								code, msg, filePath = ExecuteModel(igReq.GetModel(), igReq.GetPromptWord())
							} else {
								msg = "Unavailable ipfs storage node"
							}
						}
						timestamp := time.Now()
						if code == 0 {
							cid, code, err = ipfs.UploadImage(ctx, HttpUrl2Multiaddr(ipfsAddr), filePath)
							if err != nil {
								msg = err.Error()
								log.Logger.Errorf("Failed to upload image %s to ipfs endpoint %s", filePath, ipfsAddr)
							} else {
								_ = ipfs.WriteMFSHistory(timestamp.Unix(), ipfsAddr,
									igReq.GetModel(), igReq.GetPromptWord(),
									cid, filepath.Base(filePath))
							}
						}
						_ = db.WriteModelHistory(timestamp.Unix(), code, msg,
							igReq.GetModel(), igReq.GetPromptWord(),
							ipfsAddr, cid, filePath)
						igBody := &protocol.ImageGenerationBody{
							Data: &protocol.ImageGenerationBody_Res{
								Res: &protocol.ImageGenerationResponse{
									IpfsNode:  ipfsAddr,
									Cid:       cid,
									ImageName: filepath.Base(filePath),
								},
							},
						}
						resBody, err := proto.Marshal(igBody)
						if err != nil {
							log.Logger.Warnf("Marshal Image Generation Response Body %v", err)
							break
						}
						resBody, err = p2p.Encrypt(pmsg.Header.NodeId, resBody)
						res := protocol.Message{
							Header: &protocol.MessageHeader{
								ClientVersion: p2p.Hio.UserAgent,
								Timestamp:     timestamp.Unix(),
								Id:            pmsg.Header.Id,
								NodeId:        config.GC.Identity.PeerID,
								Receiver:      pmsg.Header.NodeId,
							},
							Type:          protocol.MessageType_IMAGE_GENERATION,
							Body:          resBody,
							ResultCode:    int32(code),
							ResultMessage: msg,
						}
						if err == nil {
							res.Header.NodePubKey, _ = p2p.MarshalPubKeyFromPrivKey(p2p.Hio.PrivKey)
						}
						resBytes, err := proto.Marshal(&res)
						if err != nil {
							log.Logger.Errorf("Marshal Image Generation Response %v", err)
							break
						}
						publishChan <- resBytes
						log.Logger.Info("Sending Image Generation Response")
					} else {
						log.Logger.Info("Gossip Image Generation Request of ", igReq.GetNodeId())
					}
				} else if igRes := ig.GetRes(); igRes != nil {
					res := serve.ImageGenerationResponse{
						Code:    int(pmsg.ResultCode),
						Message: pmsg.ResultMessage,
					}
					if pmsg.ResultCode == 0 {
						res.Data.IpfsNode = igRes.IpfsNode
						res.Data.CID = igRes.Cid
						res.Data.ImageName = igRes.ImageName
					}
					notifyData, err := json.Marshal(res)
					if err != nil {
						log.Logger.Warnf("Marshal Image Generation Response %v", err)
						break
					}
					serve.WriteAndDeleteRequestItem(pmsg.Header.Id, notifyData)
				}
			} else {
				log.Logger.Warn("Message type and body do not match")
			}
		case protocol.MessageType_HOST_INFO:
			hi := &protocol.HostInfoBody{}
			if pmsg.ResultCode != 0 {
				res := serve.HostInfoResponse{
					Code:    int(pmsg.ResultCode),
					Message: pmsg.ResultMessage,
				}
				notifyData, err := json.Marshal(res)
				if err != nil {
					log.Logger.Warnf("Marshal Identity Protocol %v", err)
					break
				}
				serve.WriteAndDeleteRequestItem(pmsg.Header.Id, notifyData)
			} else if err := proto.Unmarshal(msgBody, hi); err == nil {
				if hiReq := hi.GetReq(); hiReq != nil {
					if hiReq.GetNodeId() == config.GC.Identity.PeerID {
						hostInfo, err := host.GetHostInfo()
						var code int32 = 0
						var msg string = "ok"
						if err != nil {
							code = int32(types.ErrCodeHostInfo)
							msg = err.Error()
						}
						hiRes := &protocol.HostInfoResponse{
							Os: &protocol.HostInfoResponse_OSInfo{
								Os:              hostInfo.Os.OS,
								Platform:        hostInfo.Os.Platform,
								PlatformFamily:  hostInfo.Os.PlatformFamily,
								PlatformVersion: hostInfo.Os.PlatformVersion,
								KernelVersion:   hostInfo.Os.KernelVersion,
								KernelArch:      hostInfo.Os.KernelArch,
							},
							Memory: &protocol.HostInfoResponse_MemoryInfo{
								TotalPhysicalBytes: hostInfo.Memory.TotalPhysicalBytes,
								TotalUsableBytes:   hostInfo.Memory.TotalUsableBytes,
							},
						}
						for _, cpu := range hostInfo.Cpu {
							hiRes.Cpu = append(hiRes.Cpu, &protocol.HostInfoResponse_CpuInfo{
								ModelName:    cpu.ModelName,
								TotalCores:   cpu.Cores,
								TotalThreads: cpu.Threads,
							})
						}
						for _, disk := range hostInfo.Disk {
							hiRes.Disk = append(hiRes.Disk, &protocol.HostInfoResponse_DiskInfo{
								DriveType:    disk.DriveType,
								SizeBytes:    disk.SizeBytes,
								Model:        disk.Model,
								SerialNumber: disk.SerialNumber,
							})
						}
						for _, gpu := range hostInfo.Gpu {
							hiRes.Gpu = append(hiRes.Gpu, &protocol.HostInfoResponse_GpuInfo{
								Vendor:  gpu.Vendor,
								Product: gpu.Product,
							})
						}
						hiBody := &protocol.HostInfoBody{
							Data: &protocol.HostInfoBody_Res{
								Res: hiRes,
							},
						}
						resBody, err := proto.Marshal(hiBody)
						if err != nil {
							log.Logger.Warnf("Marshal HostInfo Response Body %v", err)
							break
						}
						resBody, err = p2p.Encrypt(pmsg.Header.NodeId, resBody)
						res := protocol.Message{
							Header: &protocol.MessageHeader{
								ClientVersion: p2p.Hio.UserAgent,
								Timestamp:     time.Now().Unix(),
								Id:            pmsg.Header.Id,
								NodeId:        config.GC.Identity.PeerID,
								Receiver:      pmsg.Header.NodeId,
							},
							Type:          protocol.MessageType_HOST_INFO,
							Body:          resBody,
							ResultCode:    code,
							ResultMessage: msg,
						}
						if err == nil {
							res.Header.NodePubKey, _ = p2p.MarshalPubKeyFromPrivKey(p2p.Hio.PrivKey)
						}
						resBytes, err := proto.Marshal(&res)
						if err != nil {
							log.Logger.Errorf("Marshal HostInfo Response %v", err)
							break
						}
						publishChan <- resBytes
						log.Logger.Info("Sending HostInfo Response")
					} else {
						log.Logger.Warnf("Invalid node id %v in request body", hiReq.GetNodeId())
					}
				} else if hiRes := hi.GetRes(); hiRes != nil {
					res := serve.HostInfoResponse{
						Code:    int(pmsg.ResultCode),
						Message: pmsg.ResultMessage,
						Data: host.HostInfo{
							Os: host.OSInfo{
								OS:              hiRes.Os.Os,
								Platform:        hiRes.Os.Platform,
								PlatformFamily:  hiRes.Os.PlatformFamily,
								PlatformVersion: hiRes.Os.PlatformVersion,
								KernelVersion:   hiRes.Os.KernelVersion,
								KernelArch:      hiRes.Os.KernelArch,
							},
							Memory: host.MemoryInfo{
								TotalPhysicalBytes: hiRes.Memory.TotalPhysicalBytes,
								TotalUsableBytes:   hiRes.Memory.TotalUsableBytes,
							},
						},
					}
					for _, cpu := range hiRes.Cpu {
						res.Data.Cpu = append(res.Data.Cpu, host.CpuInfo{
							ModelName: cpu.ModelName,
							Cores:     cpu.TotalCores,
							Threads:   cpu.TotalThreads,
						})
					}
					for _, disk := range hiRes.Disk {
						res.Data.Disk = append(res.Data.Disk, host.DiskInfo{
							DriveType:    disk.DriveType,
							SizeBytes:    disk.SizeBytes,
							Model:        disk.Model,
							SerialNumber: disk.SerialNumber,
						})
					}
					for _, gpu := range hiRes.Gpu {
						res.Data.Gpu = append(res.Data.Gpu, host.GpuInfo{
							Vendor:  gpu.Vendor,
							Product: gpu.Product,
						})
					}
					notifyData, err := json.Marshal(res)
					if err != nil {
						log.Logger.Warnf("Marshal HostInfo Protocol %v", err)
						break
					}
					serve.WriteAndDeleteRequestItem(pmsg.Header.Id, notifyData)
				}
			} else {
				log.Logger.Warn("Message type and body do not match")
			}
		default:
			res := protocol.Message{
				Header: &protocol.MessageHeader{
					ClientVersion: p2p.Hio.UserAgent,
					Timestamp:     time.Now().Unix(),
					Id:            pmsg.Header.Id,
					NodeId:        config.GC.Identity.PeerID,
					Receiver:      pmsg.Header.NodeId,
				},
				Type:          pmsg.Type,
				Body:          nil,
				ResultCode:    int32(types.ErrCodeUnsupported),
				ResultMessage: MsgNotSupported,
			}
			resBytes, err := proto.Marshal(&res)
			if err != nil {
				log.Logger.Errorf("Marshal Image Generation Response %v", err)
				break
			}
			publishChan <- resBytes
			log.Logger.Warnf("Unknowned message type", pmsg.Type)
		}
	}
}

func ExecuteModel(model, prompt string) (int, string, string) {
	if config.GC.App.ModelAPI == "" {
		return int(types.ErrCodeModel), "Model API configuration is empty", ""
	}
	req := ModelRequest{
		Prompt: prompt,
	}
	jsonData, err := json.Marshal(req)
	if err != nil {
		log.Logger.Errorf("Marshal json request error when execute model %s", model)
		return int(types.ErrCodeModel), "Marshal model request error", ""
	}
	resp, err := http.Post(
		fmt.Sprintf("%s/models/%s", config.GC.App.ModelAPI, model),
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil || resp.StatusCode != 200 {
		log.Logger.Errorf("Send model %s request error %v and status code", model, err, resp.StatusCode)
		return int(types.ErrCodeModel), fmt.Sprintf("Send request status code %d", resp.StatusCode), ""
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Logger.Errorf("Read model %s http response %v", model, err)
		return int(types.ErrCodeModel), "Read model response error", ""
	}
	response := ModelResponse{}
	if err := json.Unmarshal(body, &response); err != nil {
		log.Logger.Errorf("Unmarshal json response error when execute model %s", model)
		return int(types.ErrCodeModel), "Unmarshal model response error", ""
	}
	return response.Code, response.Status, response.ImageUrl
}

func HttpUrl2Multiaddr(url string) string {
	urlParts := strings.Split(url, "//")
	if len(urlParts) != 2 {
		return ""
	}
	url = urlParts[1]

	urlParts = strings.Split(url, "/")
	if len(urlParts) < 1 {
		return ""
	}
	url = urlParts[0]

	urlParts = strings.Split(url, ":")
	if len(urlParts) != 2 {
		return ""
	}
	return fmt.Sprintf("/ip4/%s/tcp/%s", urlParts[0], urlParts[1])
}
