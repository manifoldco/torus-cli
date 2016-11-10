package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/manifoldco/torus-cli/tools/expect/cmd"
	"github.com/manifoldco/torus-cli/tools/expect/framework"
	"github.com/manifoldco/torus-cli/tools/expect/utils"
)

const executable = "torus"

var testSuiteFlag = flag.String("suite", "default", "group of tests to run")

func main() {
	flag.Parse()

	output := framework.Output{}
	output.Log("Running tests for executable: " + executable)

	utils.Init()
	suites, err := cmd.Init()
	if err != nil {
		teardownAndExit(err, &output)
		return
	}
	err = cmd.Execute(*suites, *testSuiteFlag, executable)
	if err != nil {
		teardownAndExit(err, &output)
		return
	}

	output.Title("Complete.")
}

func teardownAndExit(err error, output *framework.Output) {
	if err.Error() != "Commands not found." {
		cmd.Teardown(executable, output)
	}
	exitError(err)
}

func exitError(err error) {
	fmt.Println("")
	fmt.Println(err.Error())
	os.Exit(1)
}
