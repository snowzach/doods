package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"sync"
	"time"

	config "github.com/spf13/viper"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"

	"github.com/snowzach/doods/conf"
	"github.com/snowzach/doods/odrpc"
)

func main() {

	l, _ := zap.NewDevelopment()
	logger := l.Sugar()

	dialOptions := []grpc.DialOption{
		grpc.WithBlock(),
	}
	if config.GetBool("server.tls") {
		dialOptions = append(dialOptions, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{InsecureSkipVerify: true})))
	} else {
		dialOptions = append(dialOptions, grpc.WithInsecure())
	}

	// Set up a connection to the gRPC server.
	conn, err := grpc.DialContext(conf.Stop.Context, net.JoinHostPort(config.GetString("server.host"), config.GetString("server.port")), dialOptions...)
	if err != nil {
		logger.Fatalw("Could not connect", "error", err)
	}
	defer conn.Close()

	// gRPC version Client
	client := odrpc.NewOdrpcClient(conn)

	// Create the request
	img, err := ioutil.ReadFile("grace_hopper.ppm")
	if err != nil {
		logger.Fatalf("Could not load grace_hopper.bmp: %v", err)
	}

	// Authentication information - ignored if not requried
	ctx := metadata.AppendToOutgoingContext(context.Background(), odrpc.DoodsAuthKeyHeader, "test123")
	// Open Stream
	stream, err := client.DetectStream(ctx)
	if err != nil {
		logger.Fatalw("Could not stream", "error", err)
	}

	start := time.Now()
	var wg sync.WaitGroup
	doneSend := make(chan struct{})

	// Send requests
	go func() {
		for x := 0; x < 200 && !conf.Stop.Bool(); x++ {
			wg.Add(1)
			request := &odrpc.DetectRequest{
				Id:           fmt.Sprintf("%d", x),
				DetectorName: "edgetpu",
				Data:         img,
				Detect: map[string]float32{
					"*": 50, //
				},
			}
			if err := stream.Send(request); err != nil {
				logger.Fatalf("could not stream send %v", err)
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
				logger.Fatalf("can not receive %v", err)
			}
			wg.Done()
			logger.Infow("Processed", "response", response)
		}
	}()

	// Wait until done sending and done receiving then close the stream
	<-doneSend
	wg.Wait()
	if err := stream.CloseSend(); err != nil {
		logger.Error(err.Error())
	}
	logger.Infow("Done", "took", time.Since(start).Seconds())
	zap.L().Sync() // Flush the logger
}
