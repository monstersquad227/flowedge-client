package service

import (
	"context"
	"flowedge-client/utils"
	pb "github.com/monstersquad227/flowedge-proto"
	"google.golang.org/grpc"
	"io"
	"log"
	"net/http"
	"time"
)

func StartAgent(serverAddr string) {
	agentID := utils.GetAgentID()
	for {
		log.Printf("Connecting to %s ...", serverAddr)
		err := connectAndServe(serverAddr, agentID)
		if err != nil {
			log.Printf("Connection failed: %v. Reconnecting in 5s...", err)
			time.Sleep(5 * time.Second)
		}
	}
}

func connectAndServe(serverAddr, agentID string) error {
	conn, err := grpc.Dial(serverAddr, grpc.WithInsecure())
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

	// 启动心跳
	go startHeartbeat(stream, agentID)

	// 接收并处理请求
	return receiveAndHandleMessages(stream, agentID)
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

func startHeartbeat(stream pb.FlowEdge_CommunicateClient, agentID string) {
	for {
		time.Sleep(5 * time.Second)
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
		} else {
			log.Printf("Sent heartbeat for %s", agentID)
		}
	}
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
			handleExecuteRequest(stream, msg.GetExecuteRequest())
		default:
			log.Printf("Unknown message type received: %v", msg.Type)
		}
	}
}

func handleExecuteRequest(stream pb.FlowEdge_CommunicateClient, req *pb.ExecuteRequest) {
	log.Printf("Executing command: %s", req.ShellCommand)

	var output string
	var errStr string
	var exitCode int32 = 0

	resp, err := http.Get("http://127.0.0.1:2375/containers/json")
	if err != nil {
		errStr = err.Error()
		exitCode = 1
	} else {
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			errStr = "read body error: " + err.Error()
			exitCode = 2
		} else {
			output = string(body)
		}
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
