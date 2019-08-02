package cmd

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"

	emptypb "github.com/golang/protobuf/ptypes/empty"
	cli "github.com/spf13/cobra"
	config "github.com/spf13/viper"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/snowzach/doods/server/rpc"
)

var ()

func init() {

	rootCmd.AddCommand(&cli.Command{
		Use:   "client",
		Short: "CLI Client",
		Long:  `CLI Client`,
		Run: func(cmd *cli.Command, args []string) { // Initialize the databse

			dialOptions := []grpc.DialOption{
				grpc.WithBlock(),
			}
			if config.GetBool("server.tls") {
				dialOptions = append(dialOptions, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{InsecureSkipVerify: true})))
			} else {
				dialOptions = append(dialOptions, grpc.WithInsecure())
			}

			// Set up a connection to the gRPC server.
			conn, err := grpc.Dial(net.JoinHostPort(config.GetString("server.host"), config.GetString("server.port")), dialOptions...)
			if err != nil {
				logger.Fatalw("Could not connect", "error", err)
			}
			defer conn.Close()

			// gRPC version Client
			versionClient := rpc.NewVersionRPCClient(conn)

			// Make RPC call
			version, err := versionClient.Version(context.Background(), &emptypb.Empty{})
			if err != nil {
				logger.Fatalw("Could not call Version", "error", err)
			}

			fmt.Printf("Version: %s\n", version.Version)

			// // gRPC thing Client
			// thingClient := rpc.NewThingRPCClient(conn)

			// // Make RPC call
			// things, err := thingClient.ThingFind(context.Background(), &emptypb.Empty{})
			// if err != nil {
			// 	logger.Fatalw("Could not call ThingFind", "error", err)
			// }

			// // Pretty print it as JSON
			// b, err := json.MarshalIndent(things, "", "  ")
			// if err != nil {
			// 	logger.Fatalw("Could not convert to JSON", "error", err)
			// }
			// fmt.Println(string(b))

			zap.L().Sync() // Flush the logger

		},
	})
}
