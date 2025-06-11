package service

import (
	"context"
	"crypto/tls"
	"errors"
	"flowedge-client/utils"
	"fmt"
	pb "github.com/monstersquad227/flowedge-proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"io"
	"log"
	"time"
)

func StartAgent(serverAddr string, tlsConfig *tls.Config) {
	agentID := utils.GetAgentID()
	for {
		log.Printf("Connecting to %s ...", serverAddr)
		err := connectAndServe(serverAddr, agentID, tlsConfig)
		if err != nil {
			log.Printf("Connection failed: %v. Reconnecting in 5s...", err)
			time.Sleep(5 * time.Second)
		}
	}
}

func connectAndServe(serverAddr, agentID string, tlsConfig *tls.Config) error {
	//conn, err := grpc.Dial(serverAddr, grpc.WithInsecure())
	conn, err := grpc.Dial(serverAddr, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
	if err != nil {
		return err
	}
	defer conn.Close()

	client := pb.NewFlowEdgeClient(conn)
	stream, err := client.Communicate(context.Background())
	if err != nil {
		return err
	}

	if err := sendRegister(stream, agentID); err != nil {
		return err
	}

	// 控制心跳生命周期
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go startHeartbeat(ctx, stream, agentID, done)

	err = receiveAndHandleMessages(stream, agentID)

	cancel()
	<-done // 等心跳 goroutine 退出后再 return

	// 启动心跳
	//go startHeartbeat(stream, agentID)

	// 接收并处理请求
	return err
	//return receiveAndHandleMessages(stream, agentID)
}

func sendRegister(stream pb.FlowEdge_CommunicateClient, agentID string) error {
	msg := &pb.StreamMessage{
		Type: pb.MessageType_REGISTER,
		Body: &pb.StreamMessage_Register{
			Register: &pb.RegisterMessage{
				AgentId:  agentID,
				Hostname: utils.GetHostname(),
				Version:  "v0.0.1",
			},
		},
	}
	log.Printf("Sending register message: %+v", msg)
	return stream.Send(msg)
}

func startHeartbeat(ctx context.Context, stream pb.FlowEdge_CommunicateClient, agentID string, done chan struct{}) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	defer close(done) // 心跳退出，通知外部

	for {
		select {
		case <-ctx.Done():
			log.Printf("Heartbeat context canceled for %s", agentID)
			return
		case <-ticker.C:
			msg := &pb.StreamMessage{
				Type: pb.MessageType_HEARTBEAT,
				Body: &pb.StreamMessage_Heartbeat{
					Heartbeat: &pb.HeartbeatMessage{
						AgentId:   agentID,
						Timestamp: time.Now().Unix(),
					},
				},
			}
			err := stream.Send(msg)
			if err != nil {
				log.Printf("Heartbeat error: %v", err)
				return // 连接断开，退出
			} else {
				log.Printf("Sent heartbeat for %s", agentID)
			}
		}
	}
	//for {
	//	time.Sleep(5 * time.Second)
	//	msg := &pb.StreamMessage{
	//		Type: pb.MessageType_HEARTBEAT,
	//		Body: &pb.StreamMessage_Heartbeat{
	//			Heartbeat: &pb.HeartbeatMessage{
	//				AgentId:   agentID,
	//				Timestamp: time.Now().Unix(),
	//			},
	//		},
	//	}
	//	err := stream.Send(msg)
	//	if err != nil {
	//		log.Printf("Heartbeat error: %v", err)
	//	} else {
	//		log.Printf("Sent heartbeat for %s", agentID)
	//	}
	//}
}

func receiveAndHandleMessages(stream pb.FlowEdge_CommunicateClient, agentID string) error {
	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			log.Printf("Stream closed by server")
			return nil
		}
		if err != nil {
			return err
		}

		switch msg.Type {
		case pb.MessageType_EXECUTE_REQUEST:
			handleExecuteRequest(stream, msg.GetExecuteRequest(), agentID)
		default:
			log.Printf("Unknown message type received: %v", msg.Type)
		}
	}
}

func handleExecuteRequest(stream pb.FlowEdge_CommunicateClient, req *pb.ExecuteRequest, agentID string) {
	log.Printf("Executing command: %s", req.Command)

	var output string
	var errStr string
	var exitCode int32 = 0

	switch req.Command {
	case "containerList":
		result, err := ListContainers()
		if err != nil {
			errStr = err.Error()
			exitCode = 1
		} else {
			output = result
		}
	case "containerCreate":
		result, err := CreateContainers(req.Image)
		if err != nil {
			errStr = err.Error()
			exitCode = 1
		} else {
			output = result
		}
	case "containerStart":
		result, err := StartContainers()
		if err != nil {
			errStr = err.Error()
			exitCode = 1
		} else {
			output = result
		}
	case "containerStop":
		result, err := StopContainer(req.ContainerId)
		if err != nil {
			errStr = err.Error()
			exitCode = 1
		} else {
			output = result
		}
	case "containerRemove":
		result, err := RemoveContainer(req.ContainerId)
		if err != nil {
			errStr = err.Error()
			exitCode = 1
		} else {
			output = result
		}
	case "containerDragon":
		result, err := containerDragon(req.Image, agentID)
		if err != nil {
			errStr = err.Error()
			exitCode = 1
		} else {
			output = result
		}
	case "imagePull":
		result, err := PullImage(req.Image)
		if err != nil {
			errStr = err.Error()
			exitCode = 1
		} else {
			output = result
		}
	case "imageList":
		result, err := ListImage()
		if err != nil {
			errStr = err.Error()
			exitCode = 1
		} else {
			output = result
		}
	default:
		log.Printf("Unknown command: %s", req.Command)
		errStr = errors.New(fmt.Sprintf("Unknown command: %s", req.Command)).Error()
	}
	sendExecuteResponse(stream, req.CommandId, exitCode, output, errStr)
}

func sendExecuteResponse(stream pb.FlowEdge_CommunicateClient, commandID string, exitCode int32, output, errStr string) {
	resp := &pb.StreamMessage{
		Type: pb.MessageType_EXECUTE_RESPONSE,
		Body: &pb.StreamMessage_ExecuteResponse{
			ExecuteResponse: &pb.ExecuteResponse{
				CommandId: commandID,
				ExitCode:  exitCode,
				Output:    output,
				Error:     errStr,
			},
		},
	}
	if err := stream.Send(resp); err != nil {
		log.Printf("Failed to send execute response: %v", err)
	} else {
		log.Printf("Sent execute response for %s", commandID)
	}
}
