// cli_contract_test.go verifies the contract between CLI flag definitions and
// everything that depends on them: documentation, scripts, CI configs.
// Tests use kong.New to introspect the grammar without running any command.
//
// Helpers parsedCLI / findCommand / findFlag are defined in cli_test.go.
package main

import (
	"testing"
)

// ── default value constants ───────────────────────────────────────────────────

// Prevents: defaultServePort constant drifting out of sync with the kong tag on
// ServeCommand.Port. If they diverge, resolveDevAddr picks one value while
// --help advertises another, and CI healthchecks probe the wrong port.
func TestCLIContract_DefaultServePortConstantMatchesFlagDefault(t *testing.T) {
	k := parsedCLI(t)
	n := findCommand(k, "serve")
	if n == nil {
		t.Fatal("command \"serve\" not found")
	}
	f := findFlag(n, "port")
	if f == nil {
		t.Fatal("serve --port not found")
	}
	if f.Default != "9080" {
		t.Errorf("serve --port kong default = %q, want %q", f.Default, "9080")
	}
	if defaultServePort != 9080 {
		t.Errorf("defaultServePort constant = %d, want 9080 — keep in sync with the --port flag default", defaultServePort)
	}
}

// ── defaults ─────────────────────────────────────────────────────────────────

// Prevents: changing the default config file name from "n8go-docs.yaml".
// Projects that omit --config rely on this default; changing it silently
// breaks every project that has not pinned the flag explicitly.
func TestCLIContract_ConfigFlagDefaultIsN8goDocsYaml(t *testing.T) {
	k := parsedCLI(t)
	f := findFlag(k.Model.Node, "config")
	if f == nil {
		t.Fatal("global flag --config not found")
	}
	if f.Default != "n8go-docs.yaml" {
		t.Errorf("--config default: got %q, want %q — changing default config name breaks zero-config projects", f.Default, "n8go-docs.yaml")
	}
}

// Prevents: enabling --json by default.
// CI scripts that check exit codes rely on plain-text output unless they
// explicitly pass --json; a true default would break those scripts silently.
func TestCLIContract_JSONFlagDefaultIsFalse(t *testing.T) {
	k := parsedCLI(t)
	f := findFlag(k.Model.Node, "json")
	if f == nil {
		t.Fatal("global flag --json not found")
	}
	if f.Default != "" && f.Default != "false" {
		t.Errorf("--json default: got %q, want false — it must be opt-in", f.Default)
	}
}

// Prevents: enabling --verbose by default on the build command.
func TestCLIContract_BuildVerboseFlagDefaultIsFalse(t *testing.T) {
	k := parsedCLI(t)
	n := findCommand(k, "build")
	if n == nil {
		t.Fatal("command \"build\" not found")
	}
	f := findFlag(n, "verbose")
	if f == nil {
		t.Fatal("build --verbose not found")
	}
	if f.Default != "" && f.Default != "false" {
		t.Errorf("build --verbose default: got %q, want false", f.Default)
	}
}

// Prevents: enabling --verbose by default on the serve command.
func TestCLIContract_ServeVerboseFlagDefaultIsFalse(t *testing.T) {
	k := parsedCLI(t)
	n := findCommand(k, "serve")
	if n == nil {
		t.Fatal("command \"serve\" not found")
	}
	f := findFlag(n, "verbose")
	if f == nil {
		t.Fatal("serve --verbose not found")
	}
	if f.Default != "" && f.Default != "false" {
		t.Errorf("serve --verbose default: got %q, want false", f.Default)
	}
}

// ── help text completeness ────────────────────────────────────────────────────

// Prevents: adding a flag on any subcommand and forgetting to write help text.
// cli_test.go checks global flags; this test covers all subcommand flags so
// every flag is visible in --help output regardless of where it is defined.
func TestCLIContract_AllSubcommandFlagsHaveHelpText(t *testing.T) {
	k := parsedCLI(t)
	for _, child := range k.Model.Children {
		for _, f := range child.Flags {
			if f.Hidden || f.Name == "help" {
				continue
			}
			if f.Help == "" {
				t.Errorf("flag --%s on command %q has no help text; flags without help are invisible in --help output", f.Name, child.Name)
			}
		}
	}
}
