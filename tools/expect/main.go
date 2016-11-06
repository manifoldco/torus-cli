package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/manifoldco/torus-cli/tools/expect/cmd"
	"github.com/manifoldco/torus-cli/tools/expect/output"
	"github.com/manifoldco/torus-cli/tools/expect/utils"
)

const executable = "torus"

var testSuiteFlag = flag.String("suite", "default", "group of tests to run")

func main() {
	flag.Parse()

	output.Log("Running tests for executable: " + executable)

	utils.Init()
	suites := cmd.Init()
	err := cmd.Execute(suites, *testSuiteFlag, executable)
	if err != nil {
		if err.Error() != "Commands not found." {
			cmd.Teardown(executable)
		}
		exitError(err)
		return
	}

	output.Title("Complete.")
}

func exitError(err error) {
	fmt.Println("")
	fmt.Println(err.Error())
	os.Exit(1)
}
