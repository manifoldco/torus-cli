package cmd

import (
	"errors"
	"reflect"

	"github.com/manifoldco/torus-cli/tools/expect/framework"
	"github.com/manifoldco/torus-cli/tools/expect/output"
)

// AvailableSuites is a compilation of the suites available
type AvailableSuites struct {
	Default []framework.Command `alias:"default"`
}

func (a AvailableSuites) getByAlias(alias string) []framework.Command {
	value := reflect.ValueOf(a)
	for i := 0; i < value.NumField(); i++ {
		tag := value.Type().Field(i)
		suite, _ := tag.Tag.Lookup("alias")
		if suite != alias {
			return nil
		}

		return value.Field(i).Interface().([]framework.Command)
	}
	return nil
}

// Init returns all available test suites
func Init() AvailableSuites {
	testSuites := AvailableSuites{
		Default: []framework.Command{
			prefsList(),
			version(),
			daemonStatus(),
			daemonStop(),
			daemonStart(),
			signup(),
			projectCreate(),
			link(),
			keypairsList(),
			status(),
			logout(),
			login(),
			projectList(),
			orgCreate(),
			orgList(),
			envCreate(),
			envList(),
			machineCreate(),
			machineList(),
			serviceCreate(),
			serviceList(),
			teamCreate(),
			teamList(),
			policyList(),
			policyView(),
			policyAllow(),
			policyDeny(),
			policyListGenerated(),
			set(),
			setOther(),
			view(),
			unsetOther(),
			viewUnset(),
			setSpecific(),
			viewSpecific(),
			unlink(),
		},
	}
	return testSuites
}

// Execute spawns a test suite to run
func Execute(suites AvailableSuites, target, executable string) error {
	commands := suites.getByAlias(target)
	if len(commands) < 1 {
		return errors.New("Commands not found.")
	}

	for _, cmd := range commands {
		err := cmd.Execute(executable)
		if err != nil {
			output.Title("------- FAILED --------")
			return err
		}
	}

	return nil
}

// Teardown is ran when the suite finishes (successful or not)
func Teardown(executable string) error {
	output.Title("Teardown")
	unlinkCmd := unlink()
	err := unlinkCmd.Execute(executable)
	if err != nil {
		return err
	}

	return nil
}
