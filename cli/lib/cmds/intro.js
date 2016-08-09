'use strict';

var Command = require('../cli/command');

var cmd = new Command(
  'intro',
  'Take your first steps with ag',
  function () {
    var msg = '';

    msg += '\n';
    msg += '\n Ag: Share and secure secrets across your team, services, and environments';
    msg += '\n -------------------------------------------------------------------------';
    msg += '\n';
    msg += '\n 1. Pick a worthy project and \'link\' it';
    msg += '\n';
    msg += '\n   $ cd ~/repos/worthy_project';
    msg += '\n   $ ag link';
    msg += '\n';
    msg += '\n   Select your personal org or create a new one and then create your first project';
    msg += '\n';
    msg += '\n 2. Add a service to your new project';
    msg += '\n';
    msg += '\n   $ ag services:create an-api-or-something';
    msg += '\n';
    msg += '\n   While you\'re in a linked directory, you don\'t need to';
    msg += ' specify the org or project';
    msg += '\n';
    msg += '\n 3. Store and encrypt your first secret!';
    msg += '\n';
    msg += '\n   $ ag set SENDGRID_TOKEN 10very000secret000token01 -s an-api-or-something';
    msg += '\n';
    msg += '\n 4. Play it back or start a service and inject it into the environment';
    msg += '\n';
    msg += '\n   $ ag view -s an-api-or-something';
    msg += '\n   $ ag run ./bin/api -s an-api-or-something';
    msg += '\n';
    msg += '\n   Running a Node app? Look for your secret at process.env.SENDGRID_TOKEN';
    msg += '\n';
    msg += '\n 5. Share you some secrets';
    msg += '\n';
    msg += '\n   $ ag invites:send skywalker@rebel.alliance';
    msg += '\n';
    msg += '\n   Luke will receive an email with instructions';
    msg += '\n';
    msg += '\n 6. Explore and experiment';
    msg += '\n';
    msg += '\n   $ ag help';
    msg += '\n   $ ag help status';
    msg += '\n';
    msg += '\n *This is a work-in-progress preview, and as such is not yet';
    msg += ' ready for your production secrets';
    msg += '\n';

    console.log(msg);
  }
);

module.exports = cmd;
