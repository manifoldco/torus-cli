package cmd

import (
	"github.com/manifoldco/expect"
)

// Signup is a test for torus signup
func Signup(namespace string) *expect.Command {
	ctx := expect.GetContext(namespace)
	c := expect.Command{
		Spawn:   "torus signup",
		Context: ctx,
		Actions: []expect.Expect{
			expect.Expect{
				Output: "Name",
				Input:  ctx.GetValue("Name"),
			},
			expect.Expect{
				Output: "Username",
				Input:  ctx.GetValue("Username"),
			},
			expect.Expect{
				Output: "Email",
				Input:  ctx.GetValue("Email"),
			},
			expect.Expect{
				Output: "Password",
				Input:  ctx.GetValue("Password"),
			},
			expect.Expect{
				Output: "Confirm Password",
				Input:  ctx.GetValue("Password"),
			},
			expect.Expect{
				OutputList: []string{
					"You are now authenticated.",
					"Keypairs generated",
					"Signing keys signed",
					"Signing keys uploaded",
					"Encryption keys signed",
					"Your account has been created!",
				},
			},
			expect.Expect{
				Output:       "Verification code",
				RequestInput: true,
				Timeout:      expect.NewDuration(20),
			},
			expect.Expect{
				Output: "Your email is now verified.",
			},
		},
	}
	return &c
}
