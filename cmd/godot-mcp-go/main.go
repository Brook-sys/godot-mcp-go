package main

import (
	"fmt"
	"os"

	"github.com/Brook-sys/godot-mcp-go/internal/mcpserver"
)

var version = "dev"

func main() {
	if err := mcpserver.Run(version); err != nil {
		fmt.Fprintf(os.Stderr, "godot-mcp-go error: %v\n", err)
		os.Exit(1)
	}
}
