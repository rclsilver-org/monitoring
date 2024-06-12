package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/rclsilver-org/monitoring/daemon/mqtt"
	"github.com/rclsilver-org/monitoring/daemon/pkg/component"
	"github.com/rclsilver-org/monitoring/daemon/pkg/server"
	"github.com/rclsilver-org/monitoring/daemon/version"
)

var mqttCmd = &cobra.Command{
	Use:   "mqtt",
	Short: "Start the MQTT monitoring daemon",
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithCancel(cmd.Context())

		signalCh := make(chan os.Signal, 1)
		signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)

		go func() {
			signal := <-signalCh
			logrus.WithContext(ctx).Debugf("received %v signal", signal)
			cancel()
		}()

		cfg, err := mqtt.LoadConfig()
		if err != nil {
			logrus.WithContext(ctx).WithError(err).Fatal("unable to load the configuration")
		}

		s, err := server.NewServer(ctx, server.WithVerbose(verbose), server.WithTitle("mqtt"), server.WithVersion(version.VersionFull()))
		if err != nil {
			logrus.WithContext(ctx).WithError(err).Fatal("unable to initialize the server")
		}

		mqtt, err := mqtt.New(cfg, s)
		if err != nil {
			logrus.WithContext(ctx).WithError(err).Fatal("unable to initialize the MQTT component")
		}

		component.Start(ctx, mqtt)

		if err := s.Serve(ctx); err != nil {
			logrus.WithContext(ctx).WithError(err).Fatal("unable to start the HTTP server")
		}
	},
}

func init() {
	rootCmd.AddCommand(mqttCmd)
}
