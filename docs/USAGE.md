# Usage Guide

## 1. Configure the MCP server

Build or download `godot-mcp-go`, then add it to your MCP client configuration.

```json
{
  "mcpServers": {
    "godot": {
      "command": "/absolute/path/to/godot-mcp-go",
      "env": {
        "GODOT_PATH": "/absolute/path/to/godot"
      }
    }
  }
}
```

## 2. Discover projects

Ask your AI client:

```text
List Godot projects under /home/me/dev
```

This uses `list_projects`.

## 3. Inspect a project

```text
Get project info for /home/me/dev/MyGame
```

This uses `get_project_info` and reports main scene and file counts.

## 4. Create and configure projects

```text
Create a Godot project named MyGame at /home/me/dev/MyGame
Set the main scene to res://scenes/main.tscn
Read project settings
Set display/window/size/viewport_width to 1280
List all .gd files
```

These use `create_project`, `set_main_scene`, `read_project_settings`, `modify_project_settings`, and `list_project_files`.

## 5. Edit files and assets

```text
Read scripts/player.gd from my Godot project
Write a new movement script to scripts/player.gd
Create directory scenes/levels
Create a CharacterBody2D script at scripts/player.gd
Create a canvas_item shader at shaders/flash.gdshader
Rename scripts/player.gd to scripts/characters/player.gd
```

These use `read_file`, `write_file`, `create_directory`, `create_script`, `manage_shader`, and `rename_file`.

## 6. Headless scene operations

Headless tools run Godot with `--headless --path <project> --script <embedded godot_operations.gd>`.

Example prompt:

```text
Create a scene res://scenes/main.tscn with Node2D root in /home/me/dev/MyGame
```

The tool call shape is:

```json
{
  "project_path": "/home/me/dev/MyGame",
  "params": {
    "scene_path": "res://scenes/main.tscn",
    "root_node_type": "Node2D"
  }
}
```

## 7. Runtime control

Install the runtime server:

```text
Install the Godot MCP runtime server into /home/me/dev/MyGame
```

Then add `scripts/mcp_interaction_server.gd` as Autoload named `McpInteractionServer`.

### Local connection

Use this when Godot and `godot-mcp-go` are on the same machine. This is the default mode.

```bash
GODOT_MCP_RUNTIME_ADDR=127.0.0.1:9090 godot-mcp-go
```

### Remote connection with SSH tunnel

Recommended remote setup. Godot keeps listening only on `127.0.0.1`, and SSH carries the traffic securely.

```bash
ssh -L 9090:127.0.0.1:9090 user@GODOT_MACHINE_IP
GODOT_MCP_RUNTIME_ADDR=127.0.0.1:9090 godot-mcp-go
```

### Native remote TCP

Use only on trusted LAN/VPN networks. The runtime server can execute GDScript and manipulate the running game.

On the Godot machine:

```bash
GODOT_MCP_BIND_HOST=0.0.0.0 \
GODOT_MCP_BIND_PORT=9090 \
GODOT_MCP_TOKEN='change-this-long-random-token' \
godot --path /home/me/dev/MyGame
```

On the MCP machine:

```bash
GODOT_MCP_RUNTIME_ADDR=GODOT_MACHINE_IP:9090 \
GODOT_MCP_TOKEN='change-this-long-random-token' \
godot-mcp-go
```

Security checklist:

- Prefer SSH tunnel for remote access.
- Use native remote TCP only behind a firewall or VPN.
- Always set `GODOT_MCP_TOKEN` when binding to `0.0.0.0` or a LAN IP.
- Restrict inbound port `9090` to the MCP machine IP.
- Do not expose the runtime server to the public internet.

Run the game and use prompts such as:

```text
Get the scene tree from the running game
Eval this in Godot: return get_tree().current_scene.name
Get /root/Main/Player position
Set /root/Main/Player health to 100
Click at 300, 200
Take a screenshot
```

## 8. Exporting

```text
Export the project with preset Windows Desktop to /tmp/MyGame.exe
```

This uses `export_project` and requires a valid Godot export preset in `export_presets.cfg`.

## 9. Releasing

Create a semver tag:

```bash
git tag v0.1.0
git push origin v0.1.0
```

GitHub Actions runs tests and GoReleaser publishes release artifacts.
