package cmd

import (
	"fmt"
	"net"
	"os"

	"net/http"
	_ "net/http/pprof" // Import for pprof

	cli "github.com/spf13/cobra"
	config "github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/snowzach/doods/conf"
)

var (

	// Config and global logger
	configFile string
	pidFile    string
	logger     *zap.SugaredLogger

	// The Root Cli Handler
	rootCmd = &cli.Command{
		Version: conf.GitVersion,
		Use:     conf.Executable,
		PersistentPreRunE: func(cmd *cli.Command, args []string) error {
			// Create Pid File
			pidFile = config.GetString("pidfile")
			if pidFile != "" {
				file, err := os.OpenFile(pidFile, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
				if err != nil {
					return fmt.Errorf("Could not create pid file: %s Error:%v", pidFile, err)
				}
				defer file.Close()
				_, err = fmt.Fprintf(file, "%d\n", os.Getpid())
				if err != nil {
					return fmt.Errorf("Could not create pid file: %s Error:%v", pidFile, err)
				}
			}
			return nil
		},
		PersistentPostRun: func(cmd *cli.Command, args []string) {
			// Remove Pid file
			if pidFile != "" {
				os.Remove(pidFile)
			}
		},
	}
)

// Execute starts the program
func Execute() {
	// Run the program
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
	}
}

// This is the main initializer handling cli, config and log
func init() {
	// Initialize configuration
	cli.OnInitialize(initConfig, initLogger, initProfiler)
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "", "Config file")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {

	// If a config file is found, read it in.
	if configFile != "" {
		config.SetConfigFile(configFile)
		err := config.ReadInConfig()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not read config file: %s ERROR: %s\n", configFile, err.Error())
			os.Exit(1)
		}

	}
}

func initLogger() {
	conf.InitLogger()
	logger = zap.S().With("package", "cmd")
}

// Profiler can explicitly listen on address/port
func initProfiler() {
	if config.GetBool("profiler.enabled") {
		hostPort := net.JoinHostPort(config.GetString("profiler.host"), config.GetString("profiler.port"))
		go http.ListenAndServe(hostPort, nil)
		logger.Infof("Profiler enabled on http://%s", hostPort)
	}
}
