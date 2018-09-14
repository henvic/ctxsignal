package ctxsignal_test

import (
	"context"
	"fmt"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/henvic/ctxsignal"
)

func ExampleWithSignals() {
	ctx, cancel := ctxsignal.WithSignals(context.Background(), syscall.SIGUSR2)
	defer cancel()

	// Remove this Go routine to test manually with kill -USR2 PID.
	go func() {
		p, _ := os.FindProcess(os.Getpid()) // always check your errors!
		_ = p.Signal(syscall.SIGUSR2)       // don't ignore errors!
	}()

	<-ctx.Done()
	fmt.Println("Signaled!")
	// Output: Signaled!
}

func ExampleWithTermination() {
	ctx, cancel := ctxsignal.WithTermination(context.Background())
	defer cancel()

	// Remove this Go routine to test manually.
	go func() {
		p, _ := os.FindProcess(os.Getpid()) // always check your errors!
		_ = p.Signal(syscall.SIGINT)        // don't ignore errors!
	}()

	<-ctx.Done()
	fmt.Println("Interrupted!")
	// Output: Interrupted!
}

func ExampleClosed() {
	ctx, cancel := ctxsignal.WithSignals(context.Background(), syscall.SIGHUP)
	defer cancel()

	// Remove this Go routine to test manually with kill -HUP PID.
	go func() {
		p, _ := os.FindProcess(os.Getpid()) // always check your errors!
		_ = p.Signal(syscall.SIGHUP)        // don't ignore errors!
	}()

	<-ctx.Done()

	sig, err := ctxsignal.Closed(ctx)

	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Println(sig)
	// Output: hangup
}

func TestWithTermination(t *testing.T) {
	testWithTermination(t, syscall.SIGINT)
	testWithTermination(t, syscall.SIGTERM)
}

func testWithTermination(t *testing.T, sig os.Signal) {
	ctx, _ := ctxsignal.WithTermination(context.Background())

	go selfSignal(sig)

	<-ctx.Done()
}

func selfSignal(s os.Signal) {
	p, err := os.FindProcess(os.Getpid())

	if err != nil {
		panic(err)
	}

	time.Sleep(5 * time.Millisecond)
	if err := p.Signal(s); err != nil {
		panic(err)
	}
}

func TestWithSignals(t *testing.T) {
	ctx, _ := ctxsignal.WithSignals(context.Background(), syscall.SIGUSR1)

	go selfSignal(syscall.SIGUSR1)

	<-ctx.Done()

	sig, err := ctxsignal.Closed(ctx)

	if err != nil {
		t.Error("Expected context to be closed by signal")
	}

	if sig != syscall.SIGUSR1 {
		t.Errorf("Expected context to be closed by %v, got %v instead", syscall.SIGUSR1, sig)
	}
}

func TestWithSignalsCanceled(t *testing.T) {
	ctx, cancel := ctxsignal.WithSignals(context.Background(), syscall.SIGUSR1)

	go func() {
		cancel()
	}()

	<-ctx.Done()

	if _, err := ctxsignal.Closed(ctx); err == nil {
		t.Error("Expected error, got nil instead")
	}
}

func TestNoClosed(t *testing.T) {
	ctx := context.Background()

	if _, err := ctxsignal.Closed(ctx); err == nil {
		t.Error("Expected error, got nil instead")
	}
}
