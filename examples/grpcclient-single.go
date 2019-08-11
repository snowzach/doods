package main

import (
	"context"
	"fmt"
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
		fmt.Println("How to run:\n\tgrpcclient-single [source file] [doods server] [detector]")
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

	img, err := ioutil.ReadFile(sourceFile)

	request := &odrpc.DetectRequest{
		Data:         img,
		DetectorName: detector,
		Detect: map[string]float32{
			"*": 50, //
		},
	}

	// Authentication information - ignored if not requried
	ctx := metadata.AppendToOutgoingContext(context.Background(), odrpc.DoodsAuthKeyHeader, "test123")

	start := time.Now()
	var wg sync.WaitGroup
	for x := 0; x < 200; x++ {
		wg.Add(1)
		go func() {
			response, err := client.Detect(ctx, request)
			if err != nil {
				log.Printf("Error: %v", err)
			} else {
				log.Printf("Processed: %v", response)
			}
			wg.Done()
		}()
	}
	wg.Wait()
	log.Printf("Done. Took: %v", time.Since(start).Seconds())
}
