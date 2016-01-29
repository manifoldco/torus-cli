'use strict';

const log = require('../util/log').get('cmds/invite');
const CommandInterface = require('../command').CommandInterface;

const users = require('../users');
const collaborators = require('../collaborators');
const credentials = require('../credentials');
const errors = require('../errors');
const services = require('../services');

class InviteCollaborator extends CommandInterface {
  execute (argv) {
    return new Promise((resolve, reject) => {
      if (argv._.length < 1) {
        return reject(new errors.ValidationError(
          'You must supply an email-address to invite'
        ));
      }

      return users.loggedIn().then((loggedIn) => {
        if (!loggedIn) {
          log.error('You must be logged in to invite a user to collaborate.');
          return resolve(false);
        }

        return Promise.all([
          collaborators.add({ email: argv._[0] }),
          services.get()
        ]);
      }).then((results) => {
        return credentials.create({
          env: results[0].env,
          user: results[0].user,
          app: results[0].app,
          services: results[1]
        }).then((credCount) => {
          var name = results[0].user.name;
          var app = results[0].app.name;
          log.print(`${name} has been added as a collaborator to ${app}`);
          log.print(
            `${credCount} unique credentials have been generated for them`);
          resolve(true);
        });
      }).catch(reject);
    });
  }
}

module.exports= {
  key: 'invite',
  synopsis: 'Invite a user to collaborate on your application with you',
  usage: 'arigato invite <user-email-address>',
  Command: InviteCollaborator,
  example: `\tlocalhost$ arigato invite jeff@mycompany.io`
};
