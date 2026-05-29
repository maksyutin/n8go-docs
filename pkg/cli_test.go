package main

import (
	"strings"
	"testing"

	"github.com/alecthomas/kong"
)

// parsedCLI builds a kong.Kong instance from the CLI struct without running any
// command. It is safe to call from tests — no side effects.
func parsedCLI(t *testing.T) *kong.Kong {
	t.Helper()
	var c CLI
	k, err := kong.New(&c,
		kong.Name("n8go-docs"),
		kong.Exit(func(int) {}), // prevent os.Exit inside tests
	)
	if err != nil {
		t.Fatalf("kong.New: %v", err)
	}
	return k
}

// findCommand walks the top-level nodes and returns the one matching name.
func findCommand(k *kong.Kong, name string) *kong.Node {
	for _, n := range k.Model.Children {
		if n.Name == name {
			return n
		}
	}
	return nil
}

// findFlag returns the flag with the given long name on a node, or nil.
func findFlag(node *kong.Node, long string) *kong.Flag {
	for _, f := range node.Flags {
		if f.Name == long {
			return f
		}
	}
	return nil
}

// ── Top-level commands ────────────────────────────────────────────────────────

func TestCLI_CommandsExist(t *testing.T) {
	k := parsedCLI(t)
	for _, name := range []string{"build", "serve", "version"} {
		if findCommand(k, name) == nil {
			t.Errorf("command %q not found", name)
		}
	}
}

func TestCLI_BuildHasAlias(t *testing.T) {
	k := parsedCLI(t)
	n := findCommand(k, "build")
	if n == nil {
		t.Fatal("command \"build\" not found")
	}
	for _, a := range n.Aliases {
		if a == "generate" {
			return
		}
	}
	t.Error("command \"build\" is missing alias \"generate\"")
}

// ── Global flags ──────────────────────────────────────────────────────────────

func TestCLI_GlobalFlag_Config(t *testing.T) {
	k := parsedCLI(t)
	f := findFlag(k.Model.Node, "config")
	if f == nil {
		t.Fatal("global flag --config not found")
	}
	if f.Short != 'c' {
		t.Errorf("--config short flag: want 'c', got %q", f.Short)
	}
	if f.Default != "n8go-docs.yaml" {
		t.Errorf("--config default: want %q, got %q", "n8go-docs.yaml", f.Default)
	}
}

func TestCLI_GlobalFlag_JSON(t *testing.T) {
	k := parsedCLI(t)
	if findFlag(k.Model.Node, "json") == nil {
		t.Error("global flag --json not found")
	}
}

// ── build flags ───────────────────────────────────────────────────────────────

func TestCLI_Build_VerboseFlag(t *testing.T) {
	k := parsedCLI(t)
	n := findCommand(k, "build")
	if n == nil {
		t.Fatal("command \"build\" not found")
	}
	f := findFlag(n, "verbose")
	if f == nil {
		t.Fatal("build --verbose not found")
	}
	if f.Short != 'v' {
		t.Errorf("build --verbose short: want 'v', got %q", f.Short)
	}
}

// ── serve flags ───────────────────────────────────────────────────────────────

func TestCLI_Serve_PortFlag(t *testing.T) {
	k := parsedCLI(t)
	n := findCommand(k, "serve")
	if n == nil {
		t.Fatal("command \"serve\" not found")
	}
	f := findFlag(n, "port")
	if f == nil {
		t.Fatal("serve --port not found")
	}
	if f.Short != 'p' {
		t.Errorf("serve --port short: want 'p', got %q", f.Short)
	}
	if f.Default != "9080" {
		t.Errorf("serve --port default: want %q, got %q", "9080", f.Default)
	}
}

func TestCLI_Serve_VerboseFlag(t *testing.T) {
	k := parsedCLI(t)
	n := findCommand(k, "serve")
	if n == nil {
		t.Fatal("command \"serve\" not found")
	}
	if findFlag(n, "verbose") == nil {
		t.Error("serve --verbose not found")
	}
}

// ── help text sanity ──────────────────────────────────────────────────────────

func TestCLI_HelpText_NotEmpty(t *testing.T) {
	k := parsedCLI(t)
	for _, name := range []string{"build", "serve", "version"} {
		n := findCommand(k, name)
		if n == nil {
			t.Errorf("command %q not found", name)
			continue
		}
		if strings.TrimSpace(n.Help) == "" {
			t.Errorf("command %q has empty help text", name)
		}
	}
}

func TestCLI_HelpText_GlobalFlagsHaveHelp(t *testing.T) {
	k := parsedCLI(t)
	for _, f := range k.Model.Node.Flags {
		if f.Name == "help" {
			continue
		}
		if strings.TrimSpace(f.Help) == "" {
			t.Errorf("global flag --%s has empty help text", f.Name)
		}
	}
}
