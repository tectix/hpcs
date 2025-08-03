package main

import (
	"os"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/tectix/hpcs/internal/config"
	"github.com/tectix/hpcs/internal/server"
	"github.com/tectix/hpcs/pkg/logger"
)

var (
	cfgFile string
	rootCmd = &cobra.Command{
		Use:   "hpcs-server",
		Short: "High-Performance Cache System Server",
		Long:  `A distributed Redis-like cache server with consistent hashing and high performance.`,
		RunE:  runServer,
	}
)

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "configs/config.yaml", "config file path")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func runServer(cmd *cobra.Command, args []string) error {
	// Load configuration
	cfg, err := config.Load(cfgFile)
	if err != nil {
		return err
	}

	// Initialize logger
	log, err := logger.New(cfg.Logging)
	if err != nil {
		return err
	}
	defer log.Sync()

	log.Info("Starting HPCS Server",
		zap.String("version", "1.0.0"),
		zap.String("config", cfgFile),
	)

	// Create and start server
	srv := server.New(cfg, log)
	return srv.Start()
}