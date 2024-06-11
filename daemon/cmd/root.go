package cmd

import (
	"os"

	"github.com/ovh/configstore"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/rclsilver/monitoring/daemon/pkg/pid"
)

var (
	verbose bool

	configFile string

	pidFile string
	pidLock pid.ProcessLockFile
)

var rootCmd = &cobra.Command{
	Use:   "monitoring-daemon",
	Short: "monitoring-daemon server",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if verbose {
			logrus.SetLevel(logrus.DebugLevel)
		}

		if pidFile != "" {
			lock, err := pid.AcquireProcessIDLock(pidFile)
			if err != nil {
				logrus.WithError(err).Fatal("unable to write the pid file")
			}
			pidLock = lock
		}

		if configFile != "" {
			logrus.Infof("loading the configuration from the file %q", configFile)
			configstore.File(configFile)
		} else {
			logrus.Infof("loading the configuration according the %q environment variable", configstore.ConfigEnvVar)
			configstore.InitFromEnvironment()
		}
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		if pidLock != nil {
			if err := pidLock.Unlock(); err != nil {
				logrus.WithError(err).Warning("unable to delete the pid file")
			}
		}
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable to verbose mode")
	rootCmd.PersistentFlags().StringVarP(&configFile, "config-file", "c", "", "Configuration file")
	rootCmd.PersistentFlags().StringVarP(&pidFile, "pid-file", "p", "", "Write a pid file")
}
