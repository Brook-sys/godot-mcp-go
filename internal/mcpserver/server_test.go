package mcpserver

import "testing"

func TestSafeProjectPathAllowsRelativePath(t *testing.T) {
	got, err := safeProjectPath("/tmp/project", "scripts/player.gd")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "/tmp/project/scripts/player.gd"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestSafeProjectPathAcceptsResPrefix(t *testing.T) {
	got, err := safeProjectPath("/tmp/project", "res://scenes/main.tscn")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "/tmp/project/scenes/main.tscn"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestSafeProjectPathRejectsTraversal(t *testing.T) {
	_, err := safeProjectPath("/tmp/project", "../secret.txt")
	if err == nil {
		t.Fatal("expected traversal error")
	}
}

func TestMapStringAnyNil(t *testing.T) {
	got := mapStringAny(nil)
	if len(got) != 0 {
		t.Fatalf("got len %d, want 0", len(got))
	}
}

func TestExtractProjectValue(t *testing.T) {
	content := "[application]\nrun/main_scene=\"res://main.tscn\"\n"
	got := extractProjectValue(content, "run/main_scene")
	if got != "res://main.tscn" {
		t.Fatalf("got %q", got)
	}
}
