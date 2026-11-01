package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "kubeclusterflow",
	Short: "A real-time Kubernetes cluster monitoring tool",
	Long:  `kubeclusterflow is a tool for monitoring Kubernetes clusters in real-time.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
