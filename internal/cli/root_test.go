package cli

import (
	"bytes"
	"os"
	"testing"
)

func TestExecute(t *testing.T) {
	// Capture output so help text doesn't pollute test logs.
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{})

	err := Execute()
	if err != nil {
		t.Fatalf("Execute() returned error: %v", err)
	}

	// Reset for other tests
	rootCmd.SetOut(os.Stdout)
	rootCmd.SetArgs(nil)
}

func TestVersionCommand(t *testing.T) {
	// Capture the output of `bt version`.
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)

	rootCmd.SetArgs([]string{"version"})
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("bt version returned error: %v", err)
	}

	got := buf.String()
	if got == "" {
		t.Error("bt version produced no output")
	}

	want := "bt (BentoTask)"
	if !bytes.Contains([]byte(got), []byte(want)) {
		t.Errorf("bt version output = %q, want it to contain %q", got, want)
	}

	// Reset for other tests
	rootCmd.SetOut(os.Stdout)
	rootCmd.SetArgs(nil)
}

func TestRootHasGlobalFlags(t *testing.T) {
	flags := []string{"json", "quiet", "no-color", "data-dir", "verbose"}

	for _, name := range flags {
		f := rootCmd.PersistentFlags().Lookup(name)
		if f == nil {
			t.Errorf("expected global flag --%s to exist", name)
		}
	}
}

func TestRootOutputDefaultsToStdout(t *testing.T) {
	// Cobra defaults to stderr when no writer is set. We override this
	// in init() with SetOut(os.Stdout) so shell piping works correctly.
	// Verify the default writer is stdout, not stderr.
	writer := rootCmd.OutOrStdout()
	if writer == os.Stderr {
		t.Error("rootCmd output writer should NOT be stderr — piping will break")
	}
	// After init(), the writer should be os.Stdout (we set it explicitly)
	if writer != os.Stdout {
		t.Error("rootCmd output writer should be os.Stdout")
	}
}
