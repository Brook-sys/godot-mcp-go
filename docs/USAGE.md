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

## 4. Edit files

```text
Read scripts/player.gd from my Godot project
Write a new movement script to scripts/player.gd
Create directory scenes/levels
```

These use `read_file`, `write_file`, and `create_directory`.

## 5. Headless scene operations

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

## 6. Runtime control

Install the runtime server:

```text
Install the Godot MCP runtime server into /home/me/dev/MyGame
```

Then add `scripts/mcp_interaction_server.gd` as Autoload named `McpInteractionServer`.

Run the game and use prompts such as:

```text
Get the scene tree from the running game
Eval this in Godot: return get_tree().current_scene.name
Get /root/Main/Player position
Set /root/Main/Player health to 100
Click at 300, 200
Take a screenshot
```

## 7. Releasing

Create a semver tag:

```bash
git tag v0.1.0
git push origin v0.1.0
```

GitHub Actions runs tests and GoReleaser publishes release artifacts.
