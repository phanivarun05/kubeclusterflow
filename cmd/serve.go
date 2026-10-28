package cmd

import (
	//"context"
	//"fmt"
	//"os"
	//"os/signal"
	//"syscall"
	"github.com/spf13/cobra"
	//"github.com/phanivarun05/kubeclusterflow/internal/api"
	//"github.com/phanivarun05/kubeclusterflow/internal/k8s"
	//"github.com/phanivarun05/kubeclusterflow/internal/localcache"
)

var portFlag string

// Serve function
func runServe(cmd *cobra.Command, args []string) error {
	print("Initializing KubeClusterFlow Dashboard...")

	// resolve kubernetes context
	return nil
}

// cobra serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Launch the real-time Kubernetes cluster monitoring web dashboard",
	RunE:  runServe,
}
