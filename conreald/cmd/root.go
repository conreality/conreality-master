/* This is free and unencumbered software released into the public domain. */

package cmd

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/conreality/conreality-master/rpc"
	"github.com/conreality/conreality-master/server"
	"github.com/grandcat/zeroconf"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

const version = "0.0.0" // FIXME

const name = "conreality"
const host = "0.0.0.0"
const port = 5555

var configFile string
var debug bool
var verbose bool

// RootCmd describes the `conreald` command
var RootCmd = &cobra.Command{
	Use:     "conreald",
	Short:   "Conreality Master",
	Version: version,
	Run: func(cmd *cobra.Command, args []string) {
		go startRPCServer()

		meta := []string{fmt.Sprintf("version=%s", version)}
		mdnsServer, err := zeroconf.Register(name, "_conreality._tcp", "local.", port, meta, nil)
		if err != nil {
			panic(err)
		}
		defer mdnsServer.Shutdown()

		sigterm := make(chan os.Signal, 1)
		signal.Notify(sigterm, os.Interrupt, syscall.SIGTERM)

		select {
		case <-sigterm:
			fmt.Println()
		}
	},
}

func startRPCServer() {
	grpcServer := grpc.NewServer()
	rpc.RegisterChatServer(grpcServer, &server.ChatServer{})

	tcpListener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		panic(err)
	}

	err = grpcServer.Serve(tcpListener)
	if err != nil {
		panic(err)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	RootCmd.PersistentFlags().StringVarP(&configFile, "config", "C", "", "Set config file (default: $HOME/.conreality/config.yaml)")
	RootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "Enable debugging")
	RootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Be verbose")
	RootCmd.SetVersionTemplate(`Conreality Master {{printf "%s" .Version}}
`)
}

// initConfig reads the configuration file and environment variables.
func initConfig() {
	if configFile != "" {
		// Use the specified config file:
		viper.SetConfigFile(configFile)
	} else {
		// Search for config file in the current directory and under the home directory:
		viper.SetConfigName("config")
		viper.AddConfigPath(".")
		viper.AddConfigPath("$HOME/.conreality")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in:
	if err := viper.ReadInConfig(); err == nil {
		if debug {
			fmt.Println("Using config file:", viper.ConfigFileUsed())
		}
	}
}

// Execute implements the `conreald` command
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
