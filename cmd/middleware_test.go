package cmd

import (
	"testing"

	"github.com/urfave/cli"
)

func TestChain(t *testing.T) {
	t.Run("aborts on first error", func(t *testing.T) {
		firstRan := false
		secondRan := false

		expected := cli.NewExitError("error", -1)
		err := Chain(
			func(ctx *cli.Context) error {
				firstRan = true
				if secondRan {
					t.Error("Second chained func ran first")
				}
				return expected
			},
			func(ctx *cli.Context) error {
				secondRan = true
				return nil
			},
		)(&cli.Context{})

		if err != expected {
			t.Error("Chain did not return first func's error")
		}

		if !firstRan {
			t.Error("First func did not run")
		}
		if secondRan {
			t.Error("Second func was run")
		}
	})

	t.Run("runs all chained funcs", func(t *testing.T) {
		firstRan := false
		secondRan := false

		Chain(
			func(ctx *cli.Context) error {
				firstRan = true
				if secondRan {
					t.Error("Second chained func ran first")
				}
				return nil
			},
			func(ctx *cli.Context) error {
				secondRan = true
				return nil
			},
		)(&cli.Context{})

		if !(firstRan && secondRan) {
			t.Error("Both chained funcs did not run")
		}
	})
}
