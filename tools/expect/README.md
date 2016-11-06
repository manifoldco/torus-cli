## Usage

```
> make test-expect
> ./test-expect
```

### Add a new command

1. Create new test function in `cmd` directory which returns `framework.Command`
2. Add function to an existing, or create a new test suite in `cmd/cmd.go`

### Add a new test suite

1. Add field and alias to `cmd/cmd.go` "AvailableSuites"
2. Add `[]framework.Command{}` for aforementioned alias in `testSuites`
