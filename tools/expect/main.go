package main

import (
	"fmt"
	"os"

	"github.com/manifoldco/expect"

	"github.com/manifoldco/torus-cli/tools/expect/cmd"
)

// Suites is a map of test suites
var Suites map[string]expect.Suite

func main() {
	err := BootstrapContext()
	if err != nil {
		exitError(err)
		return
	}

	// Test runner
	runner := expect.NewRunner()

	// Default test suite
	runner.NewSuite("default", []*expect.Command{
		cmd.Signup("userA"),
		cmd.Signup("userB"),
	})

	// Run based on flags
	err = runner.Execute()
	exitError(err)
}

func exitError(err error) {
	if err != nil {
		fmt.Println(err.Error())
	}
	os.Exit(1)
}
