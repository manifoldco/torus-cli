package main

import (
	"os"
	"os/user"
	"path/filepath"

	"github.com/manifoldco/expect"
)

// BootstrapContext appends dummy users to a global set of contexts
func BootstrapContext() error {
	a, err := createUserContext("A", SignupData{
		Name:     "John Doe",
		Email:    "johndoe" + expect.Nonce + "@example.com",
		Username: "johndoe-" + expect.Nonce,
		Password: "password",
	})
	if err != nil {
		return err
	}
	expect.AddContext("userA", a)

	b, err := createUserContext("B", SignupData{
		Name:     "Jane Smith",
		Email:    "janemith" + expect.Nonce + "@example.com",
		Username: "janesmith-" + expect.Nonce,
		Password: "password",
	})
	if err != nil {
		return err
	}
	expect.AddContext("userB", b)

	c, err := createUserContext("C", SignupData{
		Name:     "Michael Smith",
		Email:    "mikesmith" + expect.Nonce + "@example.com",
		Username: "mikesmith-" + expect.Nonce,
		Password: "password",
	})
	if err != nil {
		return err
	}
	expect.AddContext("userC", c)

	return nil
}

// Create the directory necessary for torus to operate in
func createTorusRoot(namespace string) (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}

	torusRoot := filepath.Join(usr.HomeDir, ".torus", "expect", namespace)
	if _, err := os.Stat(torusRoot); os.IsNotExist(err) {
		err := os.MkdirAll(torusRoot, 0700)
		if err != nil {
			return "", err
		}
	}

	return torusRoot, nil
}

// Helper function to create a command context for a test user
func createUserContext(namespace string, user SignupData) (*expect.CommandContext, error) {
	root, err := createTorusRoot(namespace)
	if err != nil {
		return nil, err
	}

	env := TorusEnv{
		TorusRoot: root,
	}
	context := expect.ContextFromStruct(user, env)

	return &context, nil
}
