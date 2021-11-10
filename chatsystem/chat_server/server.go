package main

import (
	"context"
	"net"
	"flag"
	"fmt"
	"log"
	"sync"
	"google.golang.org/grpc"

	pb "google.golang.org/grpc/examples/chatsystem/chatsystem"
)

var(
	port = flag.Int("port", 10000, "The server port")
)

type GreeterServer struct{
	pb.UnimplementedGreeterServer
	mu sync.Mutex // protects RecvReply
	msgs map[string][]pb.RecvReply
}


func (s *GreeterServer) SendMassage(ctx context.Context, msg *pb.SendRequest) (*pb.SendReply, error) {
	key := msg.GetRcv()
	in := msg.GetMsg()
	s.mu.Lock()
	s.msgs[key] = append(s.msgs[key], pb.RecvReply{Msg:in})
	s.mu.Unlock()

	log.Printf("Received message %v", in)

	return &pb.SendReply{Msg:"seed successful"}, nil
}

func (s *GreeterServer) RecvMassage(client *pb.RecvRequest, stream pb.Greeter_RecvMassageServer) error {
	key := client.GetRcv()
	
	// add a mutex
	s.mu.Lock()

	rn := make([]pb.RecvReply, len(s.msgs[key]))
	copy(rn, s.msgs[key])

	s.mu.Unlock()

	if len(s.msgs[key]) == 0{
		if err := stream.Send(&pb.RecvReply{Msg:""}); err != nil {
			return err
		}
	}else{
		for _, msg := range rn{
			if err := stream.Send(&msg); err != nil {
				return err
			}
		}
		delete(s.msgs, key)
	}

	return nil
}

func newServer() *GreeterServer {
	s := &GreeterServer{msgs: make(map[string][]pb.RecvReply)}
	return s
}

func main(){
	flag.Parse()

	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	var opts []grpc.ServerOption
	
	grpcServer := grpc.NewServer(opts...)

	pb.RegisterGreeterServer(grpcServer, newServer())
	grpcServer.Serve(lis)
}