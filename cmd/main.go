package main

import (
	"k8soperation/global"
	"k8soperation/internal/bootstrap"
	"k8soperation/internal/server"
	"os"
	"time"

	"github.com/spf13/cobra"
)

type RunFunc func(configFile string) error

func NewRootCommand(run RunFunc) *cobra.Command {
	var configFile string

	rootCmd := &cobra.Command{
		Use:           "k8s-manager",
		Short:         "Start k8s-manager",
		SilenceUsage:  true,
		SilenceErrors: true,
		Example:       "k8s-manager -c configs/config.yaml",
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(configFile)
		},
	}
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "", "config file path; falls back to APP_CONFIG or configs/config.yaml")
	return rootCmd
}

func main() {
	cmd := NewRootCommand(run)
	cmd.SetArgs(os.Args[1:])
	if err := cmd.Execute(); err != nil {
		panic(err)
	}
}

func run(configFile string) error {
	if err := bootstrap.InitAll(configFile); err != nil {
		return err
	}
	defer bootstrap.FlushLoggers()

	srv := server.NewHTTPServer()
	server.ListenAndServeAsync(srv)

	timeout := time.Duration(global.ServerSetting.ShutdownTimeout) * time.Second
	if timeout <= 0 {
		timeout = 5 * time.Second
	}

	server.GracefulShutdown(srv, timeout)
	return nil
}
