package main

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/snowzach/doods/odrpc"
)

func main() {

	if len(os.Args) < 4 {
		fmt.Println("How to run:\n\tgrpcclient-stream [source file] [doods server] [detector]")
		return
	}

	// parse args
	sourceFile := os.Args[1]
	server := os.Args[2]
	detector := os.Args[3]

	dialOptions := []grpc.DialOption{
		grpc.WithBlock(),
		grpc.WithInsecure(),
	}

	// Set up a connection to the gRPC server.
	conn, err := grpc.Dial(server, dialOptions...)
	if err != nil {
		log.Fatalf("Could not connect: %v", err)
	}
	defer conn.Close()

	// gRPC version Client
	client := odrpc.NewOdrpcClient(conn)

	// Create the request
	img, err := ioutil.ReadFile(sourceFile)
	if err != nil {
		log.Fatalf("Could not load %s %v", sourceFile, err)
	}

	// Authentication information - ignored if not requried
	ctx := metadata.AppendToOutgoingContext(context.Background(), odrpc.DoodsAuthKeyHeader, "test123")
	// Open Stream
	stream, err := client.DetectStream(ctx)
	if err != nil {
		log.Fatalf("Could not stream: %v", err)
	}

	start := time.Now()
	var wg sync.WaitGroup
	doneSend := make(chan struct{})

	// Send requests
	go func() {
		for x := 0; x < 200; x++ {
			wg.Add(1)
			request := &odrpc.DetectRequest{
				Id:           fmt.Sprintf("%d", x),
				DetectorName: detector,
				Data:         img,
				Detect: map[string]float32{
					"*": 50, //
				},
			}
			if err := stream.Send(request); err != nil {
				log.Fatalf("could not stream send %v", err)
			}
		}
		close(doneSend)
	}()

	// Parse results
	go func() {
		for {
			response, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatalf("can not receive %v", err)
			}
			wg.Done()
			log.Printf("Processed: %v", response)
		}
	}()

	// Wait until done sending and done receiving then close the stream
	<-doneSend
	wg.Wait()
	if err := stream.CloseSend(); err != nil {
		log.Fatal(err.Error())
	}
	log.Printf("Done. Took: %v", time.Since(start).Seconds())
}
