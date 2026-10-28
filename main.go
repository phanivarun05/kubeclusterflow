package main

import (
	// Add a blank identifier prefix to protect them from your IDE

	"embed"
)

//go:embed all:ui/dist
var frontendDist embed.FS

func main() {
	// Your code here
}
