package utils

import (
	"runtime"
	"testing"
)

func assertEqual(t *testing.T, a interface{}, b interface{}) {
	t.Helper()
	if a != b {
		t.Fatalf("%v != %v", a, b)
	}
}

func TestGetFileName(t *testing.T) {
	assertEqual(t, "test_file", GetFileName("../a/b/c/test_file.txt"))
	assertEqual(t, "test_file", GetFileName("../a/b/test_file"))
	assertEqual(t, "test_file", GetFileName("test_file.txt"))
	assertEqual(t, "test_file", GetFileName("test_file"))
	assertEqual(t, "test_file", GetFileName("test_file.txt/"))
	assertEqual(t, "test_file", GetFileName("../:/:.:/:/::/..../test_file.txt/"))
	if runtime.GOOS == "windows" {
		assertEqual(t, "test_file", GetFileName("\\a\\b\\test_file.txt"))
	}
}

func TestStripParentDirectories(t *testing.T) {
	assertEqual(t, "test/a/b.txt", StripParentDirectories("../../test/./a/b.txt"))
	assertEqual(t, "b.txt", StripParentDirectories("./b.txt"))
	assertEqual(t, "b.txt", StripParentDirectories("b.txt"))
	if runtime.GOOS == "windows" {
		assertEqual(t, "test/a/b.txt", StripParentDirectories("..\\..\\test\\.\\a\\b.txt"))
	}
}

func TestPrettifyTitle(t *testing.T) {
	assertEqual(t, "getting started", PrettifyTitle("getting-started.md"))
	assertEqual(t, "my page", PrettifyTitle("my_page.md"))
	assertEqual(t, "index", PrettifyTitle("index.md"))
}
