package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/rclsilver/monitoring/daemon/pkg/component"
	"github.com/rclsilver/monitoring/daemon/pkg/server"
	"github.com/rclsilver/monitoring/daemon/speedtest"
	"github.com/rclsilver/monitoring/daemon/version"
)

var speedtestCmd = &cobra.Command{
	Use:   "speedtest",
	Short: "Start the speedtest monitoring daemon",
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithCancel(cmd.Context())

		signalCh := make(chan os.Signal, 1)
		signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)

		go func() {
			signal := <-signalCh
			logrus.WithContext(ctx).Debugf("received %v signal", signal)
			cancel()
		}()

		cfg, err := speedtest.LoadConfig()
		if err != nil {
			logrus.WithContext(ctx).WithError(err).Fatal("unable to load the configuration")
		}

		s, err := server.NewServer(ctx, server.WithVerbose(verbose), server.WithTitle("speedtest"), server.WithVersion(version.VersionFull()))
		if err != nil {
			logrus.WithContext(ctx).WithError(err).Fatal("unable to initialize the server")
		}

		speedtest, err := speedtest.New(cfg, s)
		if err != nil {
			logrus.WithContext(ctx).WithError(err).Fatal("unable to initialize the MQTT component")
		}

		component.Start(ctx, speedtest)

		if err := s.Serve(ctx); err != nil {
			logrus.WithContext(ctx).WithError(err).Fatal("unable to start the HTTP server")
		}
	},
}

func init() {
	rootCmd.AddCommand(speedtestCmd)
}
