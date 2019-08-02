package main

import (
	"context"
	"crypto/tls"
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

	img, err := ioutil.ReadFile("grace_hopper.ppm")

	request := &odrpc.DetectRequest{
		Data:         img,
		DetectorName: "default",
		Detect: map[string]float32{
			"*": 50, //
		},
	}

	// Authentication information - ignored if not requried
	ctx := metadata.AppendToOutgoingContext(context.Background(), odrpc.DoodsAuthKeyHeader, "test123")

	start := time.Now()
	var wg sync.WaitGroup
	for x := 0; x < 200 && !conf.Stop.Bool(); x++ {
		wg.Add(1)
		go func() {
			response, err := client.Detect(ctx, request)
			if err != nil {
				logger.Errorw("Processed", "error", err)
			} else {
				logger.Infow("Processed", "response", response)
			}
			wg.Done()
		}()
	}
	wg.Wait()
	logger.Infow("Done", "took", time.Since(start).Seconds())

	zap.L().Sync() // Flush the logger
}
