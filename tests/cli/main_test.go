package cli_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/utkkat/arkdy/internal/core"
	"github.com/utkkat/arkdy/internal/transport"
)

var consoleBinary string

func TestMain(m *testing.M) {
	os.Exit(runSuite(m))
}

func runSuite(m *testing.M) int {
	dir, err := os.MkdirTemp("", "arkdy-cli-tests-")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	defer os.RemoveAll(dir)
	consoleBinary = filepath.Join(dir, "arkdy")
	if runtime.GOOS == "windows" {
		consoleBinary += ".exe"
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	args := []string{"build", "-o", consoleBinary}
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			if setting.Key == "-race" && setting.Value == "true" {
				args = append(args, "-race")
			}
		}
	}
	args = append(args, "./cmd/arkdy")
	build := exec.CommandContext(ctx, "go", args...)
	build.Dir = filepath.Join("..", "..")
	if output, err := build.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "build console: %v\n%s", err, output)
		return 1
	}
	return m.Run()
}

type consoleOutput struct {
	mu      sync.Mutex
	buffer  bytes.Buffer
	changed chan struct{}
}

func (o *consoleOutput) Write(data []byte) (int, error) {
	o.mu.Lock()
	n, err := o.buffer.Write(data)
	o.mu.Unlock()
	select {
	case o.changed <- struct{}{}:
	default:
	}
	return n, err
}

func (o *consoleOutput) text() string {
	o.mu.Lock()
	defer o.mu.Unlock()
	return o.buffer.String()
}

func (o *consoleOutput) wait(t *testing.T, want string) string {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	for {
		text := o.text()
		if strings.Contains(text, want) {
			return text
		}
		select {
		case <-o.changed:
		case <-ctx.Done():
			t.Fatalf("waiting for %q, output:\n%s", want, text)
		}
	}
}

type console struct {
	input   io.WriteCloser
	output  *consoleOutput
	command *exec.Cmd
	done    chan struct{}
	err     error
}

func startConsole(t *testing.T, args ...string) *console {
	t.Helper()
	command := exec.Command(consoleBinary, args...)
	input, err := command.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}
	c := &console{input: input, command: command, output: &consoleOutput{changed: make(chan struct{}, 1)}, done: make(chan struct{})}
	command.Stdout, command.Stderr = c.output, c.output
	if err := command.Start(); err != nil {
		_ = input.Close()
		t.Fatal(err)
	}
	go func() {
		defer close(c.done)
		c.err = command.Wait()
	}()
	t.Cleanup(func() {
		_ = input.Close()
		c.waitDone(t)
	})
	return c
}

func (c *console) waitDone(t *testing.T) {
	t.Helper()
	select {
	case <-c.done:
		if c.err != nil {
			t.Errorf("console failed: %v\n%s", c.err, c.output.text())
		}
	case <-time.After(10 * time.Second):
		_ = c.command.Process.Kill()
		<-c.done
		t.Error("console did not shut down")
	}
}

func (c *console) address(t *testing.T) string {
	t.Helper()
	text := c.output.wait(t, "Введите текст")
	for _, line := range strings.Split(text, "\n") {
		if strings.HasPrefix(line, "Адрес: ") {
			return strings.TrimPrefix(line, "Адрес: ")
		}
	}
	t.Fatal("console did not print a dialable address")
	return ""
}

func (c *console) line(t *testing.T, line string) {
	t.Helper()
	if _, err := fmt.Fprintln(c.input, line); err != nil {
		t.Fatal(err)
	}
}

func TestConsoleConversation(t *testing.T) {
	bob := startConsole(t)
	address := bob.address(t)
	alice := startConsole(t, "-peer", address)
	alice.output.wait(t, "Собеседник:")
	alice.line(t, "Привет из первого окна 👋")
	bob.output.wait(t, "] Привет из первого окна 👋")
	alice.output.wait(t, "[доставлено] Привет из первого окна 👋")
	bob.line(t, "Ответ из второго окна")
	alice.output.wait(t, "] Ответ из второго окна")
	bob.output.wait(t, "[доставлено] Ответ из второго окна")
	alice.line(t, "/connect bad-address")
	alice.output.wait(t, "Ошибка подключения:")
	alice.line(t, "после ошибки")
	bob.output.wait(t, "] после ошибки")
	alice.line(t, "/peers")
	alice.output.wait(t, "(выбран)")
	alice.line(t, "/quit")
	alice.waitDone(t)
	if runtime.GOOS == "windows" {
		_ = bob.input.Close()
	} else if err := bob.command.Process.Signal(os.Interrupt); err != nil {
		t.Fatal(err)
	}
	bob.waitDone(t)
}

func TestConsoleConnectCommandAndEOF(t *testing.T) {
	bob := startConsole(t)
	address := bob.address(t)
	alice := startConsole(t)
	alice.output.wait(t, "Введите текст")
	alice.line(t, "без собеседника")
	alice.output.wait(t, "Сначала выполните /connect")
	alice.line(t, "/connect "+address)
	alice.output.wait(t, "Собеседник:")
	alice.line(t, "connected")
	bob.output.wait(t, "] connected")
	alice.output.wait(t, "[доставлено] connected")
	_ = alice.input.Close()
	alice.waitDone(t)
}

func TestConsoleFlags(t *testing.T) {
	for _, args := range [][]string{{"-unknown"}, {"unexpected"}, {"-listen", "bad"}, {"-h"}} {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		output, err := exec.CommandContext(ctx, consoleBinary, args...).CombinedOutput()
		cancel()
		if args[0] == "-h" {
			if err != nil || !strings.Contains(string(output), "-peer") {
				t.Fatalf("help: %v, %s", err, output)
			}
		} else if err == nil {
			t.Fatalf("accepted invalid args %v", args)
		}
	}
}

func TestTerminalText(t *testing.T) {
	bob := startConsole(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	alice, err := core.NewClient(ctx, transport.HostConfig{ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"}})
	if err != nil {
		t.Fatal(err)
	}
	defer alice.Close()
	target, err := alice.Connect(ctx, bob.address(t))
	if err != nil {
		t.Fatal(err)
	}
	if err := alice.Send(ctx, target, "привет\x1b[2J\r\n\u202e"); err != nil {
		t.Fatal(err)
	}
	output := bob.output.wait(t, "] привет�[2J���")
	if strings.ContainsAny(output, "\x1b\r\u202e") {
		t.Fatalf("unsafe terminal output: %q", output)
	}
}
