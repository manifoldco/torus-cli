# Travis

The sshkey contained in `id_rsa.enc` has been uploaded to Travis CI. At
run-time they decrypt the ssh key on their end and then insert it into their
copy of ssh-agent.

This allows us to pull in our private repositories from git.
