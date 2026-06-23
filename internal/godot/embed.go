package godot

import (
	_ "embed"
	"errors"
	"os"
	"path/filepath"
)

//go:embed scripts/godot_operations.gd
var godotOperationsScript string

//go:embed scripts/mcp_interaction_server.gd
var interactionServerScript string

func InteractionServerScript() string {
	return interactionServerScript
}

func ExtractOperationsScript() (string, error) {
	dir, err := os.MkdirTemp("", "godot-mcp-go-*")
	if err != nil {
		return "", err
	}
	path := filepath.Join(dir, "godot_operations.gd")
	if err := os.WriteFile(path, []byte(godotOperationsScript), 0o600); err != nil {
		return "", errors.Join(err, os.RemoveAll(dir))
	}
	return path, nil
}
