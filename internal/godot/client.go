package godot

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const DefaultRuntimeAddress = "127.0.0.1:9090"

type Client struct {
	GodotPath      string
	RuntimeAddress string
	RuntimeToken   string
	Timeout        time.Duration
}

func NewClient() Client {
	return Client{
		GodotPath:      findGodotPath(),
		RuntimeAddress: envOrDefault("GODOT_MCP_RUNTIME_ADDR", DefaultRuntimeAddress),
		RuntimeToken:   os.Getenv("GODOT_MCP_TOKEN"),
		Timeout:        30 * time.Second,
	}
}

func (c Client) RunGodot(ctx context.Context, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, c.GodotPath, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return string(out), fmt.Errorf("godot command failed: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return string(out), nil
}

func (c Client) RunOperation(ctx context.Context, projectPath, operation string, params map[string]any) (string, error) {
	if projectPath == "" {
		return "", errors.New("project_path is required")
	}
	if _, err := os.Stat(filepath.Join(projectPath, "project.godot")); err != nil {
		return "", fmt.Errorf("invalid Godot project path: %w", err)
	}
	scriptPath, err := ExtractOperationsScript()
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(filepath.Dir(scriptPath))

	payload, err := json.Marshal(params)
	if err != nil {
		return "", err
	}
	args := []string{"--headless", "--path", projectPath, "--script", scriptPath, operation, string(payload)}
	if os.Getenv("DEBUG") == "true" || os.Getenv("GODOT_MCP_DEBUG") == "true" {
		args = append(args, "--debug-godot")
	}
	return c.RunGodot(ctx, args...)
}

func (c Client) RuntimeCommand(ctx context.Context, command string, params map[string]any) (map[string]any, error) {
	message := map[string]any{"command": command, "params": params}
	if c.RuntimeToken != "" {
		message["token"] = c.RuntimeToken
	}
	payload, err := json.Marshal(message)
	if err != nil {
		return nil, err
	}
	dialer := net.Dialer{Timeout: c.Timeout}
	conn, err := dialer.DialContext(ctx, "tcp", c.RuntimeAddress)
	if err != nil {
		return nil, fmt.Errorf("could not connect to Godot runtime at %s: %w", c.RuntimeAddress, err)
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(c.Timeout))
	if _, err := conn.Write(append(payload, '\n')); err != nil {
		return nil, err
	}
	line, err := bufio.NewReader(conn).ReadBytes('\n')
	if err != nil {
		return nil, err
	}
	var result map[string]any
	if err := json.Unmarshal(line, &result); err != nil {
		return nil, fmt.Errorf("invalid runtime JSON response: %w: %s", err, strings.TrimSpace(string(line)))
	}
	return result, nil
}

func findGodotPath() string {
	if value := os.Getenv("GODOT_PATH"); value != "" {
		return value
	}
	candidates := []string{"godot", "godot4"}
	if runtime.GOOS == "windows" {
		candidates = []string{"godot.exe", "godot4.exe"}
	}
	for _, candidate := range candidates {
		if path, err := exec.LookPath(candidate); err == nil {
			return path
		}
	}
	return candidates[0]
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
