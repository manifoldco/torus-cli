# Account
The Torus CLI can be used to manage your session and profile.

## signup
###### Added [v0.1.0](https://github.com/manifoldco/torus-cli/blob/master/CHANGELOG.md)

`torus signup` is used to create a new Torus account.

The user will be prompted to enter their name, username, email, and password (twice).

After signup the user will be sent an email which contains a verification code, this code must then be pasted into the Verification Code prompt that occurs after signup. If this prompt is aborted, the user can use the verify command to complete verification.

## verify
###### Added [v0.1.0](https://github.com/manifoldco/torus-cli/blob/master/CHANGELOG.md)

`torus verify [code]` is used to complete verification of your email address. The code is used in association with the currently logged-in user to complete the verification process.

You will not be able to manage your Torus organization until you have verified your email.

## login
###### Added [v0.1.0](https://github.com/manifoldco/torus-cli/blob/master/CHANGELOG.md)

`torus login` enables you to log into your account. Without a session you cannot interact with your Torus organization. Login prompts for your email address and password.

If you have forgotten your password please contact [support@torus.sh](mailto:support@torus.sh). At this time you cannot willingly reset a forgotten password.

## logout
###### Added [v0.1.0](https://github.com/manifoldco/torus-cli/blob/master/CHANGELOG.md)

`torus logout` will destroy your current session, after doing so you must login again before performing any further actions within your organization.

## profile
Your profile contains your name, email and password inside Torus.  

### update
###### Added [v0.17.0](https://github.com/manifoldco/torus-cli/blob/master/CHANGELOG.md)

`torus profile update` enables you to modify the authenticated user’s name, email or password. 

Currently accounts can only have one email attached to them. In the event of an email change, you will need to re-verify your account. 

### view
###### Added [v0.17.0](https://github.com/manifoldco/torus-cli/blob/master/CHANGELOG.md)

`torus profile view` displays the authenticated user’s profile information.
