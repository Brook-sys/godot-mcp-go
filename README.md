# Godot MCP Go

`godot-mcp-go` is a Go port of the Godot MCP server. It provides a single native binary that lets MCP clients control Godot 4.x projects through headless editor operations and runtime TCP commands.

The goal is the same purpose as the original TypeScript project: expose Godot project management, scene manipulation, file I/O, runtime inspection, input simulation, code evaluation, audio/physics/rendering helpers, and debugging tools to AI assistants that support MCP.

## Why Go

- Single binary: no Node.js or npm install required.
- Fast startup and low memory footprint.
- Easy release distribution for Linux, macOS, and Windows.
- Native process, filesystem, and TCP handling.

## Current status

This repository contains the production-ready Go foundation and compatibility bridge:

- MCP stdio server.
- Embedded original GDScript runtime/headless scripts.
- Godot executable discovery through `GODOT_PATH` or `PATH`.
- Runtime TCP client for `game_*` commands.
- Headless Godot operation runner.
- File I/O and project metadata tools.
- Release automation with GoReleaser.

Most runtime tools are registered and forwarded to the embedded `mcp_interaction_server.gd`. Headless operations supported by the embedded `godot_operations.gd` are also forwarded.

## Requirements

- Godot Engine 4.x available as `godot`, `godot4`, or via `GODOT_PATH`.
- An MCP-compatible client: Claude Code, Cursor, Cline, Windsurf, etc.

## Install

Download a release binary from GitHub Releases, or build locally:

```bash
go build -o godot-mcp-go ./cmd/godot-mcp-go
```

## MCP configuration

### Claude Code / Cursor / Cline

```json
{
  "mcpServers": {
    "godot": {
      "command": "/absolute/path/to/godot-mcp-go",
      "args": [],
      "env": {
        "GODOT_PATH": "/absolute/path/to/godot"
      }
    }
  }
}
```

`GODOT_PATH` is optional if Godot is already in `PATH`.

## Runtime setup

Runtime tools such as `game_eval`, `game_get_scene_tree`, `game_get_property`, `game_click`, and `game_screenshot` require a Godot autoload server inside your game.

Use the MCP tool:

```text
install_runtime_server(project_path: "/path/to/project")
```

Then in Godot:

1. Open Project Settings.
2. Go to Autoload.
3. Add `scripts/mcp_interaction_server.gd`.
4. Name it `McpInteractionServer`.
5. Run the game.

The runtime server listens on `127.0.0.1:9090` by default.

Override address for the Go MCP process:

```bash
GODOT_MCP_RUNTIME_ADDR=127.0.0.1:9090
```

## Core tools

### Project/process

- `get_godot_version`
- `launch_editor`
- `run_project`
- `stop_project`
- `get_debug_output`
- `list_projects`
- `get_project_info`
- `install_runtime_server`

### File I/O

- `read_file`
- `write_file`
- `delete_file`
- `create_directory`

### Headless Godot operations

- `create_scene`
- `add_node`
- `load_sprite`
- `export_mesh_library`
- `save_scene`
- `get_uid`
- `update_project_uids`
- `read_scene`
- `modify_scene_node`
- `remove_scene_node`
- `attach_script`
- `create_resource`
- `manage_resource`
- `manage_scene_signals`
- `manage_theme_resource`
- `manage_scene_structure`

These tools accept:

```json
{
  "project_path": "/absolute/path/to/project",
  "params": {
    "scene_path": "res://scenes/main.tscn"
  }
}
```

### Runtime tools

The `game_*` tools forward JSON directly to Godot's runtime TCP server. Examples:

```json
{
  "params": {
    "code": "return get_tree().current_scene.name"
  }
}
```

```json
{
  "params": {
    "node_path": "/root/Main/Player",
    "property": "position"
  }
}
```

## Development

```bash
go test ./...
go build ./cmd/godot-mcp-go
```

## Release

Releases are automated by GitHub Actions and GoReleaser.

Create and push a tag:

```bash
git tag v0.1.0
git push origin v0.1.0
```

The workflow builds binaries for Linux, macOS, and Windows and publishes them to the GitHub release.

## License

MIT

## Credits

This project embeds and interoperates with the GDScript bridge from `tugcantopaloglu/godot-mcp`, itself based on the original work by Coding-Solo.
