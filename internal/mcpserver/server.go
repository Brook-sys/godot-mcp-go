package mcpserver

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/renatogalera/godot-mcp-go/internal/godot"
)

type Server struct {
	client godot.Client

	mu          sync.Mutex
	runningCmd  *exec.Cmd
	debugOutput *bytes.Buffer
}

func Run(version string) error {
	app := &Server{client: godot.NewClient(), debugOutput: &bytes.Buffer{}}
	s := server.NewMCPServer(
		"godot-mcp-go",
		version,
		server.WithToolCapabilities(false),
	)
	app.registerTools(s)
	return server.ServeStdio(s)
}

func (s *Server) registerTools(m *server.MCPServer) {
	m.AddTool(mcp.NewTool("get_godot_version", mcp.WithDescription("Get installed Godot version")), s.getGodotVersion)
	m.AddTool(mcp.NewTool("launch_editor", mcp.WithDescription("Launch Godot editor for a project"), mcp.WithString("project_path", mcp.Required(), mcp.Description("Absolute path to Godot project"))), s.launchEditor)
	m.AddTool(mcp.NewTool("run_project", mcp.WithDescription("Run a Godot project"), mcp.WithString("project_path", mcp.Required(), mcp.Description("Absolute path to Godot project"))), s.runProject)
	m.AddTool(mcp.NewTool("stop_project", mcp.WithDescription("Stop the running Godot project")), s.stopProject)
	m.AddTool(mcp.NewTool("get_debug_output", mcp.WithDescription("Get captured Godot process output")), s.getDebugOutput)
	m.AddTool(mcp.NewTool("install_runtime_server", mcp.WithDescription("Copy MCP runtime autoload script into a Godot project"), mcp.WithString("project_path", mcp.Required(), mcp.Description("Absolute path to Godot project")), mcp.WithString("script_path", mcp.Description("Destination path relative to project root"))), s.installRuntimeServer)
	m.AddTool(mcp.NewTool("list_projects", mcp.WithDescription("Find Godot projects below a directory"), mcp.WithString("directory", mcp.Required(), mcp.Description("Directory to scan"))), s.listProjects)
	m.AddTool(mcp.NewTool("get_project_info", mcp.WithDescription("Get Godot project metadata"), mcp.WithString("project_path", mcp.Required(), mcp.Description("Absolute path to Godot project"))), s.getProjectInfo)
	m.AddTool(mcp.NewTool("create_project", mcp.WithDescription("Create a new Godot project"), mcp.WithString("project_path", mcp.Required(), mcp.Description("Directory for the new Godot project")), mcp.WithString("name", mcp.Required(), mcp.Description("Project name")), mcp.WithString("main_scene", mcp.Description("Optional main scene path"))), s.createProject)
	m.AddTool(mcp.NewTool("read_project_settings", mcp.WithDescription("Read project.godot as JSON"), mcp.WithString("project_path", mcp.Required(), mcp.Description("Absolute path to Godot project"))), s.readProjectSettings)
	m.AddTool(mcp.NewTool("modify_project_settings", mcp.WithDescription("Set a project.godot value"), mcp.WithString("project_path", mcp.Required(), mcp.Description("Absolute path to Godot project")), mcp.WithString("section", mcp.Required(), mcp.Description("Settings section")), mcp.WithString("key", mcp.Required(), mcp.Description("Setting key")), mcp.WithString("value", mcp.Required(), mcp.Description("Raw Godot setting value"))), s.modifyProjectSettings)
	m.AddTool(mcp.NewTool("set_main_scene", mcp.WithDescription("Set application main scene"), mcp.WithString("project_path", mcp.Required(), mcp.Description("Absolute path to Godot project")), mcp.WithString("scene_path", mcp.Required(), mcp.Description("Scene path, usually res://..."))), s.setMainScene)
	m.AddTool(mcp.NewTool("list_project_files", mcp.WithDescription("List project files with optional extension filter"), mcp.WithString("project_path", mcp.Required(), mcp.Description("Absolute path to Godot project")), mcp.WithString("extension", mcp.Description("Optional extension like .gd or gd"))), s.listProjectFiles)
	m.AddTool(mcp.NewTool("read_file", mcp.WithDescription("Read a text file inside a Godot project"), mcp.WithString("project_path", mcp.Required(), mcp.Description("Absolute path to Godot project")), mcp.WithString("file_path", mcp.Required(), mcp.Description("Path relative to project root"))), s.readFile)
	m.AddTool(mcp.NewTool("write_file", mcp.WithDescription("Create or overwrite a text file inside a Godot project"), mcp.WithString("project_path", mcp.Required(), mcp.Description("Absolute path to Godot project")), mcp.WithString("file_path", mcp.Required(), mcp.Description("Path relative to project root")), mcp.WithString("content", mcp.Required(), mcp.Description("File content"))), s.writeFile)
	m.AddTool(mcp.NewTool("create_directory", mcp.WithDescription("Create a directory inside a Godot project"), mcp.WithString("project_path", mcp.Required(), mcp.Description("Absolute path to Godot project")), mcp.WithString("directory_path", mcp.Required(), mcp.Description("Path relative to project root"))), s.createDirectory)
	m.AddTool(mcp.NewTool("delete_file", mcp.WithDescription("Delete a file inside a Godot project"), mcp.WithString("project_path", mcp.Required(), mcp.Description("Absolute path to Godot project")), mcp.WithString("file_path", mcp.Required(), mcp.Description("Path relative to project root"))), s.deleteFile)
	m.AddTool(mcp.NewTool("rename_file", mcp.WithDescription("Rename or move a file inside a Godot project"), mcp.WithString("project_path", mcp.Required(), mcp.Description("Absolute path to Godot project")), mcp.WithString("from_path", mcp.Required(), mcp.Description("Current path relative to project root")), mcp.WithString("to_path", mcp.Required(), mcp.Description("New path relative to project root"))), s.renameFile)
	m.AddTool(mcp.NewTool("create_script", mcp.WithDescription("Create a GDScript file from a template"), mcp.WithString("project_path", mcp.Required(), mcp.Description("Absolute path to Godot project")), mcp.WithString("file_path", mcp.Required(), mcp.Description("Script path relative to project root")), mcp.WithString("class_name", mcp.Description("Optional class_name")), mcp.WithString("extends", mcp.Description("Base Godot class, defaults to Node"))), s.createScript)
	m.AddTool(mcp.NewTool("manage_shader", mcp.WithDescription("Create or read a .gdshader file"), mcp.WithString("project_path", mcp.Required(), mcp.Description("Absolute path to Godot project")), mcp.WithString("action", mcp.Required(), mcp.Description("read or create")), mcp.WithString("file_path", mcp.Required(), mcp.Description("Shader path relative to project root")), mcp.WithString("shader_type", mcp.Description("canvas_item, spatial, particles, or sky")), mcp.WithString("content", mcp.Description("Shader content for create"))), s.manageShader)
	m.AddTool(mcp.NewTool("export_project", mcp.WithDescription("Run Godot export for a preset"), mcp.WithString("project_path", mcp.Required(), mcp.Description("Absolute path to Godot project")), mcp.WithString("preset", mcp.Required(), mcp.Description("Export preset name")), mcp.WithString("output_path", mcp.Required(), mcp.Description("Export output path")), mcp.WithBoolean("headless", mcp.Description("Use --headless, defaults true"))), s.exportProject)

	for _, tool := range headlessTools() {
		name := tool.name
		desc := tool.desc
		op := tool.op
		m.AddTool(mcp.NewTool(name, append([]mcp.ToolOption{mcp.WithDescription(desc), mcp.WithString("project_path", mcp.Required(), mcp.Description("Absolute path to Godot project")), mcp.WithObject("params", mcp.Description("Operation parameters passed to Godot"))}, tool.options...)...), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return s.runHeadless(ctx, req, op)
		})
	}

	for _, tool := range runtimeTools() {
		name := tool.name
		desc := tool.desc
		cmd := tool.cmd
		m.AddTool(mcp.NewTool(name, mcp.WithDescription(desc), mcp.WithObject("params", mcp.Description("Runtime command parameters"))), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return s.runRuntime(ctx, req, cmd)
		})
	}
}

type mappedTool struct {
	name    string
	desc    string
	op      string
	cmd     string
	options []mcp.ToolOption
}

func headlessTools() []mappedTool {
	return []mappedTool{
		{name: "create_scene", desc: "Create a scene with a root node", op: "create_scene"},
		{name: "add_node", desc: "Add a node to a scene", op: "add_node"},
		{name: "load_sprite", desc: "Load texture into a sprite node", op: "load_sprite"},
		{name: "export_mesh_library", desc: "Export scene as MeshLibrary", op: "export_mesh_library"},
		{name: "save_scene", desc: "Save or duplicate a scene", op: "save_scene"},
		{name: "get_uid", desc: "Get UID for a Godot resource", op: "get_uid"},
		{name: "update_project_uids", desc: "Resave resources to generate UIDs", op: "resave_resources"},
		{name: "read_scene", desc: "Read scene tree as JSON", op: "read_scene"},
		{name: "modify_scene_node", desc: "Modify node properties in a scene", op: "modify_node"},
		{name: "remove_scene_node", desc: "Remove a node from a scene", op: "remove_node"},
		{name: "attach_script", desc: "Attach script to scene node", op: "attach_script"},
		{name: "create_resource", desc: "Create a .tres resource", op: "create_resource"},
		{name: "manage_resource", desc: "Read or modify resources", op: "manage_resource"},
		{name: "manage_scene_signals", desc: "Manage scene signal connections", op: "manage_scene_signals"},
		{name: "manage_theme_resource", desc: "Manage Theme resources", op: "manage_theme_resource"},
		{name: "manage_scene_structure", desc: "Rename, duplicate, or move scene nodes", op: "manage_scene_structure"},
	}
}

func runtimeTools() []mappedTool {
	pairs := []mappedTool{
		{name: "game_screenshot", desc: "Capture runtime screenshot", cmd: "screenshot"},
		{name: "game_click", desc: "Click at a screen position", cmd: "click"},
		{name: "game_key_press", desc: "Send key press or action", cmd: "key_press"},
		{name: "game_mouse_move", desc: "Move the mouse", cmd: "mouse_move"},
		{name: "game_get_ui", desc: "Get visible UI elements", cmd: "get_ui_elements"},
		{name: "game_get_scene_tree", desc: "Get runtime scene tree", cmd: "get_scene_tree"},
		{name: "game_eval", desc: "Execute GDScript in running game", cmd: "eval"},
		{name: "game_get_property", desc: "Get node property", cmd: "get_property"},
		{name: "game_set_property", desc: "Set node property", cmd: "set_property"},
		{name: "game_call_method", desc: "Call node method", cmd: "call_method"},
		{name: "game_get_node_info", desc: "Inspect a runtime node", cmd: "get_node_info"},
		{name: "game_instantiate_scene", desc: "Instantiate a scene at runtime", cmd: "instantiate_scene"},
		{name: "game_remove_node", desc: "Remove a runtime node", cmd: "remove_node"},
		{name: "game_change_scene", desc: "Change current scene", cmd: "change_scene"},
		{name: "game_pause", desc: "Pause or unpause game", cmd: "pause"},
		{name: "game_performance", desc: "Get runtime performance metrics", cmd: "get_performance"},
		{name: "game_wait", desc: "Wait N frames", cmd: "wait"},
		{name: "game_connect_signal", desc: "Connect runtime signal", cmd: "connect_signal"},
		{name: "game_disconnect_signal", desc: "Disconnect runtime signal", cmd: "disconnect_signal"},
		{name: "game_emit_signal", desc: "Emit runtime signal", cmd: "emit_signal"},
		{name: "game_play_animation", desc: "Control AnimationPlayer", cmd: "play_animation"},
		{name: "game_tween_property", desc: "Tween node property", cmd: "tween_property"},
		{name: "game_get_nodes_in_group", desc: "Find nodes by group", cmd: "get_nodes_in_group"},
		{name: "game_find_nodes_by_class", desc: "Find nodes by class", cmd: "find_nodes_by_class"},
		{name: "game_reparent_node", desc: "Reparent a node", cmd: "reparent_node"},
	}
	advanced := map[string]string{
		"game_key_hold": "key_hold", "game_key_release": "key_release", "game_scroll": "scroll", "game_mouse_drag": "mouse_drag", "game_gamepad": "gamepad", "game_get_camera": "get_camera", "game_set_camera": "set_camera", "game_raycast": "raycast", "game_get_audio": "get_audio", "game_spawn_node": "spawn_node", "game_set_shader_param": "set_shader_param", "game_audio_play": "audio_play", "game_audio_bus": "audio_bus", "game_navigate_path": "navigate_path", "game_tilemap": "tilemap", "game_add_collision": "add_collision", "game_environment": "environment", "game_manage_group": "manage_group", "game_create_timer": "create_timer", "game_set_particles": "set_particles", "game_create_animation": "create_animation", "game_serialize_state": "serialize_state", "game_physics_body": "physics_body", "game_create_joint": "create_joint", "game_bone_pose": "bone_pose", "game_ui_theme": "ui_theme", "game_viewport": "viewport", "game_debug_draw": "debug_draw", "game_http_request": "http_request", "game_websocket": "websocket", "game_multiplayer": "multiplayer", "game_rpc": "rpc", "game_touch": "touch", "game_input_state": "input_state", "game_input_action": "input_action", "game_list_signals": "list_signals", "game_await_signal": "await_signal", "game_script": "script", "game_window": "window", "game_os_info": "os_info", "game_time_scale": "time_scale", "game_process_mode": "process_mode", "game_world_settings": "world_settings", "game_csg": "csg", "game_multimesh": "multimesh", "game_procedural_mesh": "procedural_mesh", "game_light_3d": "light_3d", "game_mesh_instance": "mesh_instance", "game_gridmap": "gridmap", "game_3d_effects": "3d_effects", "game_gi": "gi", "game_path_3d": "path_3d", "game_sky": "sky", "game_camera_attributes": "camera_attributes", "game_navigation_3d": "navigation_3d", "game_physics_3d": "physics_3d", "game_canvas": "canvas", "game_canvas_draw": "canvas_draw", "game_light_2d": "light_2d", "game_parallax": "parallax", "game_shape_2d": "shape_2d", "game_path_2d": "path_2d", "game_physics_2d": "physics_2d", "game_animation_tree": "animation_tree", "game_animation_control": "animation_control", "game_skeleton_ik": "skeleton_ik", "game_audio_effect": "audio_effect", "game_audio_bus_layout": "audio_bus_layout", "game_audio_spatial": "audio_spatial", "game_locale": "locale", "game_ui_control": "ui_control", "game_ui_text": "ui_text", "game_ui_popup": "ui_popup", "game_ui_tree": "ui_tree", "game_ui_item_list": "ui_item_list", "game_ui_tabs": "ui_tabs", "game_ui_menu": "ui_menu", "game_ui_range": "ui_range", "game_render_settings": "render_settings", "game_resource": "resource",
	}
	for name, cmd := range advanced {
		pairs = append(pairs, mappedTool{name: name, desc: strings.TrimPrefix(strings.ReplaceAll(name, "_", " "), "game "), cmd: cmd})
	}
	return pairs
}

func (s *Server) getGodotVersion(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	out, err := s.client.RunGodot(ctx, "--version")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return textResult(strings.TrimSpace(out)), nil
}

func (s *Server) runHeadless(ctx context.Context, req mcp.CallToolRequest, operation string) (*mcp.CallToolResult, error) {
	projectPath, err := req.RequireString("project_path")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := mapStringAny(req.GetArguments()["params"])
	out, err := s.client.RunOperation(ctx, projectPath, operation, params)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return textResult(strings.TrimSpace(out)), nil
}

func (s *Server) runRuntime(ctx context.Context, req mcp.CallToolRequest, command string) (*mcp.CallToolResult, error) {
	params := mapStringAny(req.GetArguments()["params"])
	result, err := s.client.RuntimeCommand(ctx, command, params)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	data, _ := json.MarshalIndent(result, "", "  ")
	return textResult(string(data)), nil
}

func (s *Server) launchEditor(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	projectPath, err := req.RequireString("project_path")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	cmd := exec.Command(s.client.GodotPath, "--path", projectPath, "--editor")
	if err := cmd.Start(); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return textResult(fmt.Sprintf("Godot editor launched with PID %d", cmd.Process.Pid)), nil
}

func (s *Server) runProject(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	projectPath, err := req.RequireString("project_path")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.runningCmd != nil && s.runningCmd.Process != nil {
		return mcp.NewToolResultError("a Godot project is already running"), nil
	}
	s.debugOutput.Reset()
	cmd := exec.Command(s.client.GodotPath, "--path", projectPath)
	cmd.Stdout = s.debugOutput
	cmd.Stderr = s.debugOutput
	if err := cmd.Start(); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	s.runningCmd = cmd
	go func() {
		_ = cmd.Wait()
		s.mu.Lock()
		if s.runningCmd == cmd {
			s.runningCmd = nil
		}
		s.mu.Unlock()
	}()
	return textResult(fmt.Sprintf("Godot project running with PID %d", cmd.Process.Pid)), nil
}

func (s *Server) stopProject(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.runningCmd == nil || s.runningCmd.Process == nil {
		return textResult("No Godot project is running"), nil
	}
	pid := s.runningCmd.Process.Pid
	if err := s.runningCmd.Process.Kill(); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	s.runningCmd = nil
	return textResult(fmt.Sprintf("Stopped Godot project PID %d", pid)), nil
}

func (s *Server) getDebugOutput(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return textResult(s.debugOutput.String()), nil
}

func (s *Server) installRuntimeServer(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	projectPath, err := req.RequireString("project_path")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	dest := stringArg(req, "script_path", "scripts/mcp_interaction_server.gd")
	path, err := safeProjectPath(projectPath, dest)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	if err := os.WriteFile(path, []byte(godot.InteractionServerScript()), 0o644); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return textResult(fmt.Sprintf("Runtime server installed at %s. Add it as Autoload named McpInteractionServer.", dest)), nil
}

func (s *Server) listProjects(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	dir, err := req.RequireString("directory")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	var projects []string
	_ = filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil || !d.IsDir() {
			return nil
		}
		if _, statErr := os.Stat(filepath.Join(path, "project.godot")); statErr == nil {
			projects = append(projects, path)
			return filepath.SkipDir
		}
		if strings.HasPrefix(d.Name(), ".") && path != dir {
			return filepath.SkipDir
		}
		return nil
	})
	data, _ := json.MarshalIndent(projects, "", "  ")
	return textResult(string(data)), nil
}

func (s *Server) getProjectInfo(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	projectPath, err := req.RequireString("project_path")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	info := map[string]any{"project_path": projectPath}
	settingsPath := filepath.Join(projectPath, "project.godot")
	settings, err := os.ReadFile(settingsPath)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	info["has_project_godot"] = true
	info["main_scene"] = extractProjectValue(string(settings), "run/main_scene")
	counts := map[string]int{}
	_ = filepath.WalkDir(projectPath, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		counts[strings.ToLower(filepath.Ext(path))]++
		return nil
	})
	info["file_counts"] = counts
	data, _ := json.MarshalIndent(info, "", "  ")
	return textResult(string(data)), nil
}

func (s *Server) createProject(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	projectPath, err := req.RequireString("project_path")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	name, err := req.RequireString("name")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	if err := os.MkdirAll(projectPath, 0o755); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	mainScene := stringArg(req, "main_scene", "")
	content := buildProjectGodot(name, mainScene)
	if err := os.WriteFile(filepath.Join(projectPath, "project.godot"), []byte(content), 0o644); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return textResult("Godot project created: " + projectPath), nil
}

func (s *Server) readProjectSettings(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	projectPath, err := req.RequireString("project_path")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	data, err := os.ReadFile(filepath.Join(projectPath, "project.godot"))
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	settings := parseProjectSettings(string(data))
	out, _ := json.MarshalIndent(settings, "", "  ")
	return textResult(string(out)), nil
}

func (s *Server) modifyProjectSettings(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	projectPath, err := req.RequireString("project_path")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	section, err := req.RequireString("section")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	key, err := req.RequireString("key")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	value, err := req.RequireString("value")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	settingsPath := filepath.Join(projectPath, "project.godot")
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	updated := setProjectSetting(string(data), section, key, value)
	if err := os.WriteFile(settingsPath, []byte(updated), 0o644); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return textResult(fmt.Sprintf("Set [%s] %s=%s", section, key, value)), nil
}

func (s *Server) setMainScene(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	scenePath, err := req.RequireString("scene_path")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	args := req.GetArguments()
	args["section"] = "application"
	args["key"] = "run/main_scene"
	args["value"] = quoteGodotString(scenePath)
	return s.modifyProjectSettings(ctx, req)
}

func (s *Server) listProjectFiles(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	projectPath, err := req.RequireString("project_path")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	extension := strings.TrimSpace(stringArg(req, "extension", ""))
	if extension != "" && !strings.HasPrefix(extension, ".") {
		extension = "." + extension
	}
	var files []string
	_ = filepath.WalkDir(projectPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if strings.HasPrefix(d.Name(), ".") && path != projectPath {
				return filepath.SkipDir
			}
			return nil
		}
		if extension != "" && !strings.EqualFold(filepath.Ext(path), extension) {
			return nil
		}
		rel, err := filepath.Rel(projectPath, path)
		if err == nil {
			files = append(files, filepath.ToSlash(rel))
		}
		return nil
	})
	out, _ := json.MarshalIndent(files, "", "  ")
	return textResult(string(out)), nil
}

func (s *Server) readFile(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path, err := requiredProjectFile(req, "file_path")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return textResult(string(data)), nil
}

func (s *Server) writeFile(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path, err := requiredProjectFile(req, "file_path")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	content, err := req.RequireString("content")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return textResult("File written: " + path), nil
}

func (s *Server) createDirectory(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	projectPath, err := req.RequireString("project_path")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	dirPath, err := req.RequireString("directory_path")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	path, err := safeProjectPath(projectPath, dirPath)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	if err := os.MkdirAll(path, 0o755); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return textResult("Directory created: " + path), nil
}

func (s *Server) deleteFile(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path, err := requiredProjectFile(req, "file_path")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	if err := os.Remove(path); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return textResult("File deleted: " + path), nil
}

func (s *Server) renameFile(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	from, err := requiredProjectFile(req, "from_path")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	to, err := requiredProjectFile(req, "to_path")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	if err := os.MkdirAll(filepath.Dir(to), 0o755); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	if err := os.Rename(from, to); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return textResult(fmt.Sprintf("Renamed %s to %s", from, to)), nil
}

func (s *Server) createScript(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path, err := requiredProjectFile(req, "file_path")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	className := stringArg(req, "class_name", "")
	extends := stringArg(req, "extends", "Node")
	var builder strings.Builder
	if className != "" {
		builder.WriteString("class_name " + className + "\n")
	}
	builder.WriteString("extends " + extends + "\n\n")
	builder.WriteString("func _ready() -> void:\n\tpass\n\n")
	builder.WriteString("func _process(delta: float) -> void:\n\tpass\n")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	if err := os.WriteFile(path, []byte(builder.String()), 0o644); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return textResult("Script created: " + path), nil
}

func (s *Server) manageShader(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	action, err := req.RequireString("action")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	path, err := requiredProjectFile(req, "file_path")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	if action == "read" {
		data, err := os.ReadFile(path)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return textResult(string(data)), nil
	}
	if action == "create" {
		shaderType := stringArg(req, "shader_type", "canvas_item")
		content := stringArg(req, "content", "shader_type "+shaderType+";\n\nfunc fragment() {\n\t// COLOR = vec4(1.0);\n}\n")
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return textResult("Shader created: " + path), nil
	}
	return mcp.NewToolResultError("invalid action: " + action), nil
}

func (s *Server) exportProject(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	projectPath, err := req.RequireString("project_path")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	preset, err := req.RequireString("preset")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	outPath, err := req.RequireString("output_path")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	headless := true
	if h, ok := req.GetArguments()["headless"].(bool); ok {
		headless = h
	}
	args := []string{"--path", projectPath, "--export-release", preset, outPath}
	if headless {
		args = append(args, "--headless")
	}
	out, err := s.client.RunGodot(ctx, args...)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return textResult(strings.TrimSpace(out)), nil
}

func textResult(text string) *mcp.CallToolResult {
	return mcp.NewToolResultText(text)
}

func requiredProjectFile(req mcp.CallToolRequest, key string) (string, error) {
	projectPath, err := req.RequireString("project_path")
	if err != nil {
		return "", err
	}
	filePath, err := req.RequireString(key)
	if err != nil {
		return "", err
	}
	return safeProjectPath(projectPath, filePath)
}

func safeProjectPath(projectPath, rel string) (string, error) {
	if projectPath == "" || rel == "" {
		return "", fmt.Errorf("project path and relative path are required")
	}
	cleanProject, err := filepath.Abs(projectPath)
	if err != nil {
		return "", err
	}
	cleanRel := strings.TrimPrefix(strings.TrimPrefix(rel, "res://"), string(filepath.Separator))
	cleanPath, err := filepath.Abs(filepath.Join(cleanProject, cleanRel))
	if err != nil {
		return "", err
	}
	if cleanPath != cleanProject && !strings.HasPrefix(cleanPath, cleanProject+string(filepath.Separator)) {
		return "", fmt.Errorf("path escapes project root: %s", rel)
	}
	return cleanPath, nil
}

func stringArg(req mcp.CallToolRequest, key, fallback string) string {
	if value, ok := req.GetArguments()[key].(string); ok && value != "" {
		return value
	}
	return fallback
}

func mapStringAny(value any) map[string]any {
	if value == nil {
		return map[string]any{}
	}
	if typed, ok := value.(map[string]any); ok {
		return typed
	}
	data, _ := json.Marshal(value)
	var out map[string]any
	_ = json.Unmarshal(data, &out)
	if out == nil {
		return map[string]any{}
	}
	return out
}

func buildProjectGodot(name, mainScene string) string {
	var builder strings.Builder
	builder.WriteString("; Engine configuration file.\n")
	builder.WriteString("; Generated by godot-mcp-go.\n\n")
	builder.WriteString("config_version=5\n\n")
	builder.WriteString("[application]\n\n")
	builder.WriteString("config/name=")
	builder.WriteString(quoteGodotString(name))
	builder.WriteString("\n")
	if mainScene != "" {
		builder.WriteString("run/main_scene=")
		builder.WriteString(quoteGodotString(mainScene))
		builder.WriteString("\n")
	}
	return builder.String()
}

func parseProjectSettings(content string) map[string]map[string]string {
	settings := map[string]map[string]string{}
	section := ""
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, ";") {
			continue
		}
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			section = strings.TrimSuffix(strings.TrimPrefix(line, "["), "]")
			if settings[section] == nil {
				settings[section] = map[string]string{}
			}
			continue
		}
		if section == "" || !strings.Contains(line, "=") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		settings[section][strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
	}
	return settings
}

func setProjectSetting(content, section, key, value string) string {
	lines := strings.Split(content, "\n")
	sectionHeader := "[" + section + "]"
	inSection := false
	sectionFound := false
	keySet := false
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]") {
			if inSection && !keySet {
				before := append([]string{}, lines[:i]...)
				before = append(before, key+"="+value)
				lines = append(before, lines[i:]...)
				keySet = true
				break
			}
			inSection = trimmed == sectionHeader
			if inSection {
				sectionFound = true
			}
			continue
		}
		if inSection && strings.HasPrefix(trimmed, key+"=") {
			lines[i] = key + "=" + value
			keySet = true
			break
		}
	}
	if !sectionFound {
		if strings.TrimSpace(content) != "" && !strings.HasSuffix(content, "\n") {
			content += "\n"
		}
		return content + "\n" + sectionHeader + "\n\n" + key + "=" + value + "\n"
	}
	if !keySet {
		lines = append(lines, key+"="+value)
	}
	return strings.Join(lines, "\n")
}

func quoteGodotString(value string) string {
	escaped := strings.ReplaceAll(value, "\\", "\\\\")
	escaped = strings.ReplaceAll(escaped, "\"", "\\\"")
	return "\"" + escaped + "\""
}

func extractProjectValue(content, key string) string {
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, key+"=") {
			return strings.Trim(strings.TrimPrefix(line, key+"="), "\"")
		}
	}
	return ""
}

var _ = time.Second
