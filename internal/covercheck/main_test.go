package main

import (
	"errors"
	"testing"
)

type exitCode int

func TestMainFunction(t *testing.T) {
	origExit, origRun := exit, runGoTest
	defer func() { exit, runGoTest = origExit, origRun }()
	runGoTest = func() ([]byte, error) { return []byte("ok pkg coverage: 100.0% of statements\n"), nil }
	exit = func(code int) { panic(exitCode(code)) }
	defer func() {
		recovered := recover()
		code, ok := recovered.(exitCode)
		if !ok {
			t.Fatalf("main() panic = %#v, want exitCode", recovered)
		}
		if code != 0 {
			t.Fatalf("main() exit = %d, want 0", code)
		}
	}()
	main()
}

func TestRunCommandFailure(t *testing.T) {
	origRun := runGoTest
	defer func() { runGoTest = origRun }()
	runGoTest = func() ([]byte, error) { return []byte("boom"), errors.New("go test") }
	if code := run(); code != 1 {
		t.Fatalf("run() = %d, want 1", code)
	}
}

func TestRunCoverageFailure(t *testing.T) {
	origRun := runGoTest
	defer func() { runGoTest = origRun }()
	runGoTest = func() ([]byte, error) { return []byte("ok pkg coverage: 0.0% of statements\n"), nil }
	if code := run(); code != 1 {
		t.Fatalf("run() = %d, want 1", code)
	}
}

func TestRequireFullCoverage(t *testing.T) {
	out := []byte("ok pkg coverage: 100.0% of statements\n")
	if err := requireFullCoverage(out); err != nil {
		t.Fatalf("requireFullCoverage() error = %v", err)
	}
}

func TestRequireFullCoverageFailures(t *testing.T) {
	if err := requireFullCoverage([]byte("no coverage here")); err == nil {
		t.Fatal("requireFullCoverage() accepted missing coverage")
	}
	out := []byte("ok pkg coverage: 99.9% of statements\n")
	if err := requireFullCoverage(out); err == nil {
		t.Fatal("requireFullCoverage() accepted partial coverage")
	}
}

func TestOutLineFallback(t *testing.T) {
	if got := outLine([]byte("abc\n"), []byte("missing")); got != "missing" {
		t.Fatalf("outLine() = %q, want missing", got)
	}
}
