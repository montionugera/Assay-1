package main

import (
	"fmt"
	"io"
	"runtime"
	"runtime/debug"
	"strings"

	"github.com/use-assay/assay/internal/attest"
)

// runVersion prints what produced this binary, one `KEY<TAB>VALUE` line each:
// the scanner identity the v3 preimage binds, every preimage version the
// binary can emit, and the Go toolchain and VCS revision it was built from. The
// first two are what a verifier needs to reproduce an evidence_hash; the build
// lines are for bug reports, and are deliberately not part of any hash (see
// attest.ScannerIdentity for why a commit hash is not the scanner identity).
func runVersion(args []string, d commandDeps) error {
	if len(args) != 0 {
		return fmt.Errorf("version takes no arguments")
	}
	writeVersion(d.stdout, debug.ReadBuildInfo)
	return nil
}

// writeVersion takes the build-info reader as a parameter so tests can pin the
// output for a VCS-stamped build, which `go test` binaries never are.
func writeVersion(w io.Writer, readBuildInfo func() (*debug.BuildInfo, bool)) {
	goVersion := runtime.Version()
	revision := "unknown"
	modified := ""
	if info, ok := readBuildInfo(); ok {
		if info.GoVersion != "" {
			goVersion = info.GoVersion
		}
		for _, s := range info.Settings {
			switch s.Key {
			case "vcs.revision":
				revision = s.Value
			case "vcs.modified":
				modified = s.Value
			}
		}
	}

	preimages := strings.Join([]string{
		attest.PreimageVersion,
		attest.PreimageVersionCheckSet,
		attest.PreimageVersionNetwork,
	}, ",")

	_, _ = fmt.Fprintf(w, "scanner\t%s\n", attest.ScannerIdentity)
	_, _ = fmt.Fprintf(w, "preimage\t%s\n", preimages)
	_, _ = fmt.Fprintf(w, "go\t%s\n", goVersion)
	_, _ = fmt.Fprintf(w, "vcs.revision\t%s\n", revision)
	if modified != "" {
		_, _ = fmt.Fprintf(w, "vcs.modified\t%s\n", modified)
	}
}
