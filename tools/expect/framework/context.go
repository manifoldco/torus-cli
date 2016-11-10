package framework

import (
	"context"
	"os"
	"os/user"
	"path/filepath"
	"time"

	"github.com/manifoldco/torus-cli/tools/expect/utils"
)

// ExpectContextKey is context key namespace
type ExpectContextKey int

// ExpectContext is a context value for expect expect tests
type ExpectContext struct {
	Namespace string
	TorusRoot string
	User      *SignupData
	Timeout   *time.Duration
}

// NewUserContext creates an expect test user context
func NewUserContext(namespace string, userData *SignupData) (context.Context, error) {
	usr, err := user.Current()
	if err != nil {
		return nil, err
	}

	torusRoot := filepath.Join(usr.HomeDir, ".torus", "expect", namespace)
	if _, err := os.Stat(torusRoot); os.IsNotExist(err) {
		err := os.MkdirAll(torusRoot, 0700)
		if err != nil {
			return nil, err
		}
	}

	c := context.Background()
	ctxKey := ExpectContextKey(0)

	timeout := utils.Duration(30)

	expectCtx := ExpectContext{
		Namespace: namespace,
		TorusRoot: torusRoot,
		User:      userData,
		Timeout:   &timeout,
	}

	ctx := context.WithValue(c, ctxKey, expectCtx)
	return ctx, nil
}

// ContextValue returns the value for a context namespace
func ContextValue(c *context.Context) ExpectContext {
	ctx := *c
	key := ExpectContextKey(0)
	value := ctx.Value(key)
	return value.(ExpectContext)
}
