package platform

import (
	"os"
	"strings"
	"testing"
)

func TestRemovePathEntryRemovesExactEntryOnly(t *testing.T) {
	sep := string(os.PathListSeparator)
	target := "/tmp/bgit"
	pathValue := strings.Join([]string{
		"/usr/bin",
		target,
		"/tmp/bgit-tools",
		target,
		"/opt/bin",
	}, sep)

	got := RemovePathEntry(pathValue, target)
	want := strings.Join([]string{
		"/usr/bin",
		"/tmp/bgit-tools",
		"/opt/bin",
	}, sep)

	if got != want {
		t.Fatalf("RemovePathEntry() = %q, want %q", got, want)
	}
}

func TestRemovePathEntryKeepsUnmatchedPath(t *testing.T) {
	sep := string(os.PathListSeparator)
	pathValue := strings.Join([]string{"/usr/bin", "/opt/bin"}, sep)

	got := RemovePathEntry(pathValue, "/tmp/bgit")
	if got != pathValue {
		t.Fatalf("RemovePathEntry() = %q, want %q", got, pathValue)
	}
}
