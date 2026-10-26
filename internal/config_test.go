package internal

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/spf13/pflag"
)

func listenConfig(listen string) string {
	return fmt.Sprintf("listen: %q\n", listen)
}

// writeInPlace overwrites the file with a single write, without truncating it,
// since koanf ignores an identical event fired within 5ms (here the write that
// would follow the truncation). The content must keep the same length.
func writeInPlace(t *testing.T, path, content string) {
	t.Helper()
	f, err := os.OpenFile(path, os.O_WRONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if _, err := f.WriteAt([]byte(content), 0); err != nil {
		t.Fatal(err)
	}
}

// renameOver replaces the file by renaming a new one over it, like most
// editors do.
func renameOver(t *testing.T, path, content string) {
	t.Helper()
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(tmp, path); err != nil {
		t.Fatal(err)
	}
}

// waitFor waits until a value containing want is received.
func waitFor(t *testing.T, values <-chan string, want string) {
	t.Helper()
	timeout := time.After(5 * time.Second)
	for {
		select {
		case got := <-values:
			if strings.Contains(got, want) {
				return
			}
		case <-timeout:
			t.Fatalf("timed out waiting for %q", want)
		}
	}
}

type logWriter chan string

func (w logWriter) Write(p []byte) (int, error) {
	select {
	case w <- string(p):
	default:
	}
	return len(p), nil
}

func TestConfigWatchFileReplaced(t *testing.T) {
	logs := make(logWriter, 100)
	log.SetOutput(logs)
	t.Cleanup(func() { log.SetOutput(os.Stderr) })

	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte(listenConfig(":1")), 0644); err != nil {
		t.Fatal(err)
	}

	config := NewConfig(path, pflag.NewFlagSet("test", pflag.ContinueOnError))
	if err := config.Load(); err != nil {
		t.Fatal(err)
	}

	reloads := make(chan string, 100)
	config.Watch(func() {
		reloads <- config.Listen
	})

	writeInPlace(t, path, listenConfig(":2"))
	waitFor(t, reloads, ":2")

	// Unlink then create, like install(1) does.
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(listenConfig(":3")), 0644); err != nil {
		t.Fatal(err)
	}
	waitFor(t, reloads, ":3")

	// The watch must still be active after the file has been replaced.
	writeInPlace(t, path, listenConfig(":4"))
	waitFor(t, reloads, ":4")

	renameOver(t, path, listenConfig(":5"))
	waitFor(t, reloads, ":5")

	writeInPlace(t, path, listenConfig(":6"))
	waitFor(t, reloads, ":6")

	// An invalid configuration is ignored and a later valid one is applied.
	renameOver(t, path, "listen: [\":7\n")
	waitFor(t, logs, "keeping the previous one")
	writeInPlace(t, path, listenConfig(":7"))
	waitFor(t, reloads, ":7")
}
