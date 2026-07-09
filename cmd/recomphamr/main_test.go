package main

import (
	"os"
	"testing"
)

type exitCode int

func TestMainFunction(t *testing.T) {
	originalArgs := os.Args
	originalExit := exit
	defer func() {
		os.Args = originalArgs
		exit = originalExit
	}()

	os.Args = []string{"recomphamr", "--diagnostic"}
	exit = func(code int) {
		panic(exitCode(code))
	}

	defer func() {
		recovered := recover()
		code, ok := recovered.(exitCode)
		if !ok {
			t.Fatalf("main() panic = %#v, want exitCode", recovered)
		}
		if code != 0 {
			t.Fatalf("main() exit code = %d, want 0", code)
		}
	}()

	main()
}
