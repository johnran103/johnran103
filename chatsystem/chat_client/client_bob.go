package main

import (
	"context"
	"sync"
	"log"
	"flag"
	"fmt"
	"google.golang.org/grpc"
	"time"

	pb "google.golang.org/grpc/examples/chatsystem/chatsystem"
)

var(
	serverAddr = flag.String("server_addr", "localhost:10000", "The server address in the format of host:port")
	myName = flag.String("", "Bob", "Name of me")
)

func runSendMassage(client pb.GreeterClient, msg pb.SendRequest){
	fmt.Println("Sending message!")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	rpy, err := client.SendMassage(ctx, &msg)
	if err != nil {
		log.Fatalf("%v send failed because of %v: ", client, err)
	}
	log.Println(rpy)
}

func runRecvMassage(client pb.GreeterClient, rcvRqst pb.RecvRequest){
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	stream, err := client.RecvMassage(ctx, &rcvRqst)
	if err != nil {
		log.Fatalf("%v received failed because of %v", client, err)
	}
	for {
		msg, err := stream.Recv()
		if msg.GetMsg() == "" {
			break
		}else{
			if err != nil {
				log.Fatalf("%v received failed because of, %v", client, err)
			}
			log.Printf("%s", msg.GetMsg())
		}
	}
}

func main(){
	flag.Parse()
	var opts []grpc.DialOption

	opts = append(opts, grpc.WithInsecure())
	opts = append(opts, grpc.WithBlock())

	conn, err := grpc.Dial(*serverAddr, opts...)

	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}

	defer conn.Close()

	client := pb.NewGreeterClient(conn)

	var recv, msg string

	var wg sync.WaitGroup
	wg.Add(2)

	go func(){
		for{
			//time.Sleep(1 * time.Second)
			_, err = fmt.Scan(&recv, &msg)
			if err != nil{
				fmt.Println("read error becuase of %v", err)
			}

			runSendMassage(client, pb.SendRequest{Rcv:recv, Msg:msg})
		}
		wg.Done()
	}()

	go func(){
		for{
			runRecvMassage(client, pb.RecvRequest{Rcv:*myName})
			time.Sleep(1 * time.Second)
		}
		wg.Done()
	}()

	wg.Wait()
}