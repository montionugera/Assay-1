package main

import (
	"bytes"
	"runtime/debug"
	"strings"
	"testing"

	"github.com/use-assay/assay/internal/attest"
)

// TestVersionDispatch pins that every spelling of the version request prints
// the scanner identity and every preimage version the binary can produce, and
// nothing on stderr.
func TestVersionDispatch(t *testing.T) {
	for _, args := range [][]string{{"version"}, {"-version"}, {"--version"}} {
		d, out, errOut, _ := deps(testReport(), nil)
		if err := runWith(args, d); err != nil {
			t.Fatalf("runWith(%v) = %v", args, err)
		}
		got := out.String()
		for _, want := range []string{
			"scanner\t" + attest.ScannerIdentity + "\n",
			"preimage\t" + attest.PreimageVersion + "," + attest.PreimageVersionCheckSet + "," + attest.PreimageVersionNetwork + "\n",
			"go\t",
			"vcs.revision\t",
		} {
			if !strings.Contains(got, want) {
				t.Errorf("runWith(%v) output missing %q:\n%s", args, want, got)
			}
		}
		if errOut.Len() != 0 {
			t.Errorf("runWith(%v) wrote to stderr: %q", args, errOut.String())
		}
	}
}

// TestVersionRejectsArguments keeps `assay version` a closed command, so a
// typo like `assay version scan` is not silently accepted.
func TestVersionRejectsArguments(t *testing.T) {
	d, _, _, _ := deps(testReport(), nil)
	err := runWith([]string{"version", "extra"}, d)
	if err == nil {
		t.Fatal("runWith(version extra) = nil, want an error")
	}
	if got := exitCodeForError(err); got != exitUsage {
		t.Errorf("exit code = %d, want %d (usage)", got, exitUsage)
	}
}

// TestWriteVersionBuildInfo pins the exact output for a stamped build and the
// fallbacks for one without VCS stamping (e.g. `go test`, `go run`).
func TestWriteVersionBuildInfo(t *testing.T) {
	preimages := attest.PreimageVersion + "," + attest.PreimageVersionCheckSet + "," + attest.PreimageVersionNetwork

	t.Run("stamped", func(t *testing.T) {
		var b bytes.Buffer
		writeVersion(&b, func() (*debug.BuildInfo, bool) {
			return &debug.BuildInfo{
				GoVersion: "go1.99.1",
				Settings: []debug.BuildSetting{
					{Key: "vcs", Value: "git"},
					{Key: "vcs.revision", Value: "0123abcd"},
					{Key: "vcs.modified", Value: "true"},
				},
			}, true
		})
		want := "scanner\t" + attest.ScannerIdentity + "\n" +
			"preimage\t" + preimages + "\n" +
			"go\tgo1.99.1\n" +
			"vcs.revision\t0123abcd\n" +
			"vcs.modified\ttrue\n"
		if b.String() != want {
			t.Errorf("writeVersion =\n%s\nwant\n%s", b.String(), want)
		}
	})

	t.Run("unstamped", func(t *testing.T) {
		var b bytes.Buffer
		writeVersion(&b, func() (*debug.BuildInfo, bool) { return nil, false })
		got := b.String()
		if !strings.Contains(got, "vcs.revision\tunknown\n") {
			t.Errorf("missing unknown revision:\n%s", got)
		}
		if !strings.Contains(got, "go\tgo") {
			t.Errorf("missing runtime Go version fallback:\n%s", got)
		}
		if strings.Contains(got, "vcs.modified") {
			t.Errorf("vcs.modified printed without VCS info:\n%s", got)
		}
	})
}
