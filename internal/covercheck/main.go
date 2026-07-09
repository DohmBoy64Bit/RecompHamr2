package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
)

var coverageLine = regexp.MustCompile(`coverage: ([0-9]+(?:\.[0-9]+)?)% of statements`)
var exit = os.Exit
var runGoTest = func() ([]byte, error) {
	cmd := exec.Command("go", "test", "-cover", "./...")
	return cmd.CombinedOutput()
}

func main() {
	exit(run())
}

func run() int {
	out, err := runGoTest()
	os.Stdout.Write(out)
	if err != nil {
		return 1
	}
	if err := requireFullCoverage(out); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return 0
}

func requireFullCoverage(out []byte) error {
	matches := coverageLine.FindAllSubmatch(out, -1)
	if len(matches) == 0 {
		return fmt.Errorf("coverage check found no package coverage lines")
	}
	for _, match := range matches {
		value, err := strconv.ParseFloat(string(match[1]), 64)
		if err != nil {
			return err
		}
		if value != 100 {
			line := outLine(out, match[0])
			return fmt.Errorf("coverage below 100%%: %s", line)
		}
	}
	return nil
}

func outLine(out []byte, fragment []byte) string {
	for _, line := range bytes.Split(out, []byte{'\n'}) {
		if bytes.Contains(line, fragment) {
			return string(line)
		}
	}
	return string(fragment)
}
