package main

import (
	"os"
	"time"

	migratecmd "k8soperation/cmd/migrate"
	"k8soperation/internal/bootstrap"
	"k8soperation/internal/server"

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
	rootCmd.AddCommand(migratecmd.NewCommand(&configFile))
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
	app, err := bootstrap.InitAll(configFile)
	if err != nil {
		return err
	}
	defer bootstrap.FlushLoggers(app)

	srv := server.NewHTTPServer(app)
	server.ListenAndServeAsync(app.Logger, srv)

	timeout := time.Duration(app.ServerSetting.ShutdownTimeout) * time.Second
	if timeout <= 0 {
		timeout = 5 * time.Second
	}

	server.GracefulShutdown(app, srv, timeout)
	return nil
}
