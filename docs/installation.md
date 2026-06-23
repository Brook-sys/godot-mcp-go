# Installation

Download the correct archive from the latest GitHub Release:

```text
https://github.com/Brook-sys/godot-mcp-go/releases/latest
```

`godot-mcp-go` is an MCP stdio server. It normally does not print a CLI UI when executed directly; configure it in your MCP client after installing the binary.

## Requirements

- Godot Engine 4.x installed locally or available through `GODOT_PATH`.
- An MCP-compatible client such as Claude Code, Cursor, Cline, Windsurf, or another MCP client.

## Linux x86_64

```bash
curl -L -o godot-mcp-go.tar.gz https://github.com/Brook-sys/godot-mcp-go/releases/latest/download/godot-mcp-go_Linux_x86_64.tar.gz
tar -xzf godot-mcp-go.tar.gz
chmod +x godot-mcp-go
sudo mv godot-mcp-go /usr/local/bin/
which godot-mcp-go
```

## Linux ARM64

```bash
curl -L -o godot-mcp-go.tar.gz https://github.com/Brook-sys/godot-mcp-go/releases/latest/download/godot-mcp-go_Linux_arm64.tar.gz
tar -xzf godot-mcp-go.tar.gz
chmod +x godot-mcp-go
sudo mv godot-mcp-go /usr/local/bin/
which godot-mcp-go
```

## macOS Intel

```bash
curl -L -o godot-mcp-go.tar.gz https://github.com/Brook-sys/godot-mcp-go/releases/latest/download/godot-mcp-go_Darwin_x86_64.tar.gz
tar -xzf godot-mcp-go.tar.gz
chmod +x godot-mcp-go
sudo mv godot-mcp-go /usr/local/bin/
which godot-mcp-go
```

If macOS blocks the binary because it is unsigned:

```bash
xattr -d com.apple.quarantine /usr/local/bin/godot-mcp-go
```

## macOS Apple Silicon

```bash
curl -L -o godot-mcp-go.tar.gz https://github.com/Brook-sys/godot-mcp-go/releases/latest/download/godot-mcp-go_Darwin_arm64.tar.gz
tar -xzf godot-mcp-go.tar.gz
chmod +x godot-mcp-go
sudo mv godot-mcp-go /usr/local/bin/
which godot-mcp-go
```

If macOS blocks the binary because it is unsigned:

```bash
xattr -d com.apple.quarantine /usr/local/bin/godot-mcp-go
```

## Windows x86_64

Download:

```text
godot-mcp-go_Windows_x86_64.zip
```

Extract `godot-mcp-go.exe` and place it somewhere in your `PATH`, for example:

```text
C:\Tools\godot-mcp-go\godot-mcp-go.exe
```

Then test from PowerShell:

```powershell
Get-Command godot-mcp-go.exe
```

## Windows ARM64

Download:

```text
godot-mcp-go_Windows_arm64.zip
```

Extract `godot-mcp-go.exe` and add its folder to your `PATH`.

## Build from source

```bash
git clone https://github.com/Brook-sys/godot-mcp-go.git
cd godot-mcp-go
go build -o godot-mcp-go ./cmd/godot-mcp-go
./godot-mcp-go
```

The last command starts the MCP stdio server and waits for MCP JSON-RPC input. Stop it with `Ctrl+C` if you are running it manually.

## Install with Go

```bash
go install github.com/Brook-sys/godot-mcp-go/cmd/godot-mcp-go@latest
```

Make sure your Go bin directory is in `PATH`:

```bash
export PATH="$PATH:$(go env GOPATH)/bin"
```

## Verify checksum

Download `checksums.txt` from the same release and run:

```bash
sha256sum -c checksums.txt
```

On macOS, use:

```bash
shasum -a 256 -c checksums.txt
```

## Configure Godot path

If Godot is not available as `godot` or `godot4` in `PATH`, set `GODOT_PATH` in your MCP client configuration:

```json
{
  "mcpServers": {
    "godot": {
      "command": "/usr/local/bin/godot-mcp-go",
      "env": {
        "GODOT_PATH": "/absolute/path/to/godot"
      }
    }
  }
}
```

## MCP client configuration

### Local Godot runtime

```json
{
  "mcpServers": {
    "godot": {
      "command": "/usr/local/bin/godot-mcp-go",
      "env": {
        "GODOT_PATH": "/absolute/path/to/godot",
        "GODOT_MCP_RUNTIME_ADDR": "127.0.0.1:9090"
      }
    }
  }
}
```

### Remote Godot runtime with token

```json
{
  "mcpServers": {
    "godot": {
      "command": "/usr/local/bin/godot-mcp-go",
      "env": {
        "GODOT_PATH": "/absolute/path/to/godot",
        "GODOT_MCP_RUNTIME_ADDR": "192.168.1.50:9090",
        "GODOT_MCP_TOKEN": "change-this-long-random-token"
      }
    }
  }
}
```

Use remote native TCP only on trusted LAN/VPN networks. Prefer SSH tunnels for remote access when possible.

## Runtime server setup

Runtime tools such as `game_eval`, `game_get_scene_tree`, and `game_get_property` require the Godot autoload bridge.

After installing `godot-mcp-go`, ask your MCP client to run:

```text
install_runtime_server(project_path: "/path/to/godot/project")
```

Then in Godot:

1. Open Project Settings.
2. Go to Autoload.
3. Add `scripts/mcp_interaction_server.gd`.
4. Name it `McpInteractionServer`.
5. Run the game.
