# Account
The Torus CLI can be used to manage your session and profile.

## login
###### Added [v0.1.0](https://github.com/manifoldco/torus-cli/blob/master/CHANGELOG.md)

`torus login` enables you to log into your account. Without a session you cannot interact with your Torus organization. Login prompts for your email address and password.

For commands that need you to be logged in, you can also optionally set `TORUS_EMAIL` and `TORUS_PASSWORD` as environment variables and skip this step. This is especially useful if you're using torus as a part of a CI / CD process. For machines, you can set `TORUS_TOKEN_ID` and `TORUS_TOKEN_SECRET`. 

If you have forgotten your password please contact [support@torus.sh](mailto:support@torus.sh). At this time you cannot willingly reset a forgotten password.

## logout
###### Added [v0.1.0](https://github.com/manifoldco/torus-cli/blob/master/CHANGELOG.md)

`torus logout` will destroy your current session, after doing so you must login again before performing any further actions within your organization.

## profile
Your profile contains your name, email and password inside Torus.

### update
###### Added [v0.17.0](https://github.com/manifoldco/torus-cli/blob/master/CHANGELOG.md)

`torus profile update` enables you to modify the authenticated user’s full name, email or password.

Currently accounts can only have one email attached to them. In the event of an email change, you will need to re-verify your account.

### view
###### Added [v0.17.0](https://github.com/manifoldco/torus-cli/blob/master/CHANGELOG.md)

`torus profile view` displays the authenticated user’s profile information such as their name, email and account status.
