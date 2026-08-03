// Copyright 2026 Democratized Data Foundation
//
// This file is part of the DefraDB test suite.
//
// The DefraDB test suite is licensed under either:
//
//   (1) GNU Affero General Public License v3
//   (2) Business Source License 1.1
//
// See tests/LICENSE for details.

/*
Package external drives a DefraDB release binary as a separate OS process,
exposing it through the same testing client interface as the in-process
wrappers so the rest of the test harness can drive it unchanged.

This exists so P2P cross-version compatibility tests can run an old release
alongside a native HEAD node.
*/
package external

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"sync"
	"testing"
	"time"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/event"
	"github.com/sourcenetwork/defradb/http"
)

var _ client.TxnStore = (*Wrapper)(nil)
var _ client.P2P = (*Wrapper)(nil)

// maxTxnRetries is returned by MaxTxnRetries. There is no in-process node to
// ask, so this mirrors the default used by db.DB when nothing is configured.
const maxTxnRetries = 5

// healthCheckTimeout bounds how long NewWrapper waits for the child process
// to start answering health checks.
const healthCheckTimeout = 30 * time.Second

// startAttempts is how many times NewWrapper retries a start/health failure,
// which can happen when the chosen API port is taken before the child binds.
const startAttempts = 3

// Wrapper runs a defradb binary as a child process and drives it over HTTP,
// implementing the same client.TxnStore/client.P2P surface as the in-process
// wrappers.
//
// *http.Client is embedded so its methods satisfy client.TxnStore and
// client.P2P without hand-written forwarding; only methods needing
// process-level state (Close, Events, ...) are defined explicitly.
type Wrapper struct {
	*http.Client
	cmd     *exec.Cmd
	bus     event.Bus
	rootDir string
	apiURL  string
	stderr  *ringBuffer
	// logWG tracks the log-streaming goroutines, so Close can wait for them
	// to finish before returning (they call t.Log, which panics if called
	// after the test has completed).
	logWG *sync.WaitGroup
}

// NewWrapper starts a defradb node from binaryPath as a child process, waits
// for it to become ready, and returns a Wrapper driving it over HTTP.
//
// A temporary rootdir and a free API port are chosen internally. The P2P
// listener uses port 0 so the node picks a free port at bind time; read its
// real address back with PeerInfo. Use Host for the API URL.
func NewWrapper(ctx context.Context, t testing.TB, binaryPath string) (*Wrapper, error) {
	// The API port is chosen before start, so another process can grab it in the
	// gap before the child binds. Retry a few times on a start/health failure.
	var lastErr error
	for range startAttempts {
		if err := ctx.Err(); err != nil {
			// Prefer the last attempt's error, which carries the child's
			// stderr; fall back to the bare ctx error before any attempt.
			if lastErr != nil {
				return nil, lastErr
			}
			return nil, err
		}
		w, err := startWrapper(ctx, t, binaryPath)
		if err == nil {
			return w, nil
		}
		lastErr = err
	}
	return nil, lastErr
}

// startWrapper makes one attempt to start and reach a node.
func startWrapper(ctx context.Context, t testing.TB, binaryPath string) (*Wrapper, error) {
	apiPort, err := freePort()
	if err != nil {
		return nil, errors.Wrap("failed to find free api port", err)
	}
	rootDir, err := os.MkdirTemp("", "defradb-external-*")
	if err != nil {
		return nil, errors.Wrap("failed to create rootdir", err)
	}

	apiURL := fmt.Sprintf("127.0.0.1:%d", apiPort)

	cmd := exec.CommandContext(ctx, binaryPath, "start",
		"--url", apiURL,
		"--p2paddr", "/ip4/127.0.0.1/tcp/0",
		"--store", "badger",
		"--development",
		"--no-keyring",
		"--rootdir", rootDir,
	)

	stderr := newRingBuffer(64 * 1024)
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		removeAll(rootDir)
		return nil, errors.Wrap("failed to get stdout pipe", err)
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		removeAll(rootDir)
		return nil, errors.Wrap("failed to get stderr pipe", err)
	}
	var logWG sync.WaitGroup
	streamLogs(&logWG, t, stdoutPipe, stderr, "[external stdout] ")
	streamLogs(&logWG, t, stderrPipe, stderr, "[external stderr] ")

	if err := cmd.Start(); err != nil {
		logWG.Wait()
		removeAll(rootDir)
		return nil, errors.Wrap("failed to start process", err)
	}

	httpClient, err := http.NewClient("http://" + apiURL)
	if err != nil {
		killAndWait(cmd)
		logWG.Wait()
		removeAll(rootDir)
		return nil, errors.Wrap("failed to create http client", err)
	}

	if err := waitForHealth(ctx, httpClient, healthCheckTimeout); err != nil {
		killAndWait(cmd)
		logWG.Wait()
		removeAll(rootDir)
		return nil, errors.Wrap(
			"external node did not become healthy in time",
			err,
			errors.NewKV("stderr", stderr.String()),
		)
	}

	return &Wrapper{
		Client:  httpClient,
		cmd:     cmd,
		bus:     event.NewChannelBus(1, 1),
		rootDir: rootDir,
		apiURL:  "http://" + apiURL,
		stderr:  stderr,
		logWG:   &logWG,
	}, nil
}

// waitForHealth polls the client's health-check endpoint until it responds
// successfully, the timeout elapses, or ctx is cancelled.
func waitForHealth(ctx context.Context, c *http.Client, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	var lastErr error
	for time.Now().Before(deadline) {
		hctx, cancel := context.WithTimeout(ctx, 2*time.Second)
		err := c.HealthCheck(hctx)
		cancel()
		if err == nil {
			return nil
		}
		lastErr = err

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(500 * time.Millisecond):
		}
	}
	return lastErr
}

// freePort opens a listener on an ephemeral port, reads the port, and closes
// it. There is a small race between closing and the caller binding to it,
// which is accepted here as elsewhere in the test suite.
func freePort() (int, error) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	addr, ok := l.Addr().(*net.TCPAddr)
	if !ok {
		return 0, errors.Join(errors.New("listener address is not TCP"), l.Close())
	}
	if err := l.Close(); err != nil {
		return 0, err
	}
	return addr.Port, nil
}

// streamLogs starts a goroutine that copies lines from r to both t.Log (with
// the given prefix) and buf, so recent output is available if startup fails.
// wg is done once the goroutine exits, so callers can wait for it before the
// test that owns t completes (t.Log panics if called afterwards).
func streamLogs(
	wg *sync.WaitGroup,
	t testing.TB,
	r interface{ Read([]byte) (int, error) },
	buf *ringBuffer,
	prefix string,
) {
	wg.Go(func() {
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			line := scanner.Text()
			buf.WriteString(prefix + line + "\n")
			if t != nil {
				t.Log(prefix + line)
			}
		}
	})
}

// killAndWait terminates the child process and waits for it to exit,
// swallowing errors since this is only used on already-failing paths.
func killAndWait(cmd *exec.Cmd) {
	if cmd.Process != nil {
		_ = cmd.Process.Kill()
	}
	_ = cmd.Wait()
}

// removeAll removes dir, swallowing errors since this is only used on
// already-failing setup paths.
func removeAll(dir string) {
	_ = os.RemoveAll(dir)
}

// Host returns the base URL the wrapper's HTTP client is talking to.
func (w *Wrapper) Host() string {
	return w.apiURL
}

// Close kills the child process, waits for it to exit, and removes its
// rootdir. Errors are ignored, since Close cannot fail and these are
// best-effort cleanups of a process being killed and a temp dir.
func (w *Wrapper) Close() {
	if w.cmd.Process != nil {
		_ = w.cmd.Process.Kill()
	}
	// Wait's error is expected here since the process was just killed above.
	_ = w.cmd.Wait()
	// The pipes are only closed once Wait sees the process exit, so the log
	// goroutines may still be draining them; wait before returning to avoid
	// a t.Log call racing past the end of the test.
	w.logWG.Wait()
	_ = os.RemoveAll(w.rootDir)
	w.bus.Close()
}

// Events returns a standalone event bus owned by the wrapper. It never
// fires: there is no in-process node to observe. It exists so setup code
// (e.g. state.NewEventState) can subscribe without panicking. Cross-version
// sync assertions must be driven off the native node's bus or by polling
// this node over HTTP instead.
func (w *Wrapper) Events() event.Bus {
	return w.bus
}

// MaxTxnRetries returns a fixed default, since there is no in-process node
// to ask for its configured value.
func (w *Wrapper) MaxTxnRetries() int {
	return maxTxnRetries
}

// ringBuffer is a size-bounded buffer used to retain recent stderr/stdout
// output so it can be surfaced in an error if startup fails.
type ringBuffer struct {
	mu  sync.Mutex
	max int
	buf bytes.Buffer
}

func newRingBuffer(max int) *ringBuffer {
	return &ringBuffer{max: max}
}

func (r *ringBuffer) WriteString(s string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.buf.WriteString(s)
	if r.buf.Len() > r.max {
		excess := r.buf.Len() - r.max
		r.buf.Next(excess)
	}
}

func (r *ringBuffer) String() string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.buf.String()
}
