package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "kubeclusterflow",
	Short: "A real-time Kubernetes cluster monitoring tool",
	Long:  `kubeclusterflow is a tool for monitoring Kubernetes clusters in real-time.`,
}

func Execute() error {
	return rootCmd.Execute()
}
