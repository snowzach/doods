package cmd

import (
	cli "github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/snowzach/doods/conf"
	"github.com/snowzach/doods/detector"
	"github.com/snowzach/doods/odrpc"
	"github.com/snowzach/doods/server"
)

func init() {
	rootCmd.AddCommand(apiCmd)
}

var (
	apiCmd = &cli.Command{
		Use:   "api",
		Short: "Start API",
		Long:  `Start API`,
		Run: func(cmd *cli.Command, args []string) { // Initialize the databse

			// Create the detector mux server
			d := detector.New()

			// Create the server
			s, err := server.New()
			if err != nil {
				logger.Fatalw("Could not create server",
					"error", err,
				)
			}

			// Register the RPC server and it's GRPC Gateway for when it starts
			odrpc.RegisterOdrpcServer(s.GRPCServer(), d)
			s.GWReg(odrpc.RegisterOdrpcHandlerFromEndpoint)

			err = s.ListenAndServe()
			if err != nil {
				logger.Fatalw("Could not start server",
					"error", err,
				)
			}

			<-conf.Stop.Chan() // Wait until StopChan
			conf.Stop.Wait()   // Wait until everyone cleans up
			zap.L().Sync()     // Flush the logger

		},
	}
)
