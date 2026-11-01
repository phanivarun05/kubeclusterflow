package main

import (
	// Add a blank identifier prefix to protect them from your IDE

	"embed"

	_ "helm.sh/helm/v3/pkg/release"
)

//go:embed all:ui/dist
var frontendDist embed.FS

func main() {
	// Your code here
}
