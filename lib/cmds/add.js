'use strict';

const inquirer = require('inquirer');

const log = require('../util/log').get('cmds/add');
const CommandInterface = require('../command').CommandInterface;

const errors = require('../errors');
const users = require('../users');
const services = require('../services');

const QUESTIONS = [
  {
    type: 'input',
    name: 'username',
    message: 'Sendgrid Username'
  },
  {
    type: 'password',
    name: 'password',
    message: 'Sendgrid Password'
  }
];

class AddService extends CommandInterface {
  execute (argv) {
    return new Promise((resolve, reject) => {
      
      if (argv._.length < 1) {
        return reject(new errors.ValidationError(
          'You must supply a service to add'
        ));
      }

      var serviceName = argv._[0].toLowerCase();
      if (serviceName !== 'sendgrid') {
        return reject(new errors.ValidationError(
          'Only sendgrid is supported at this time.'
        ));
      }

      users.loggedIn().then((loggedIn) => {
        if (!loggedIn) {
          log.error('You must be logged in to add a service to an app.');
          return resolve(false);
        }

        inquirer.prompt(QUESTIONS, function(answers) {
          var opts = {
            serviceName: serviceName,
            credentials: answers
          };

          return services.add(opts).then((service) => {
            var name = service.name;
            log.print(`${name} has been added to your application.`);
            log.print(`You can now \`arigato run\` your application`);

            resolve(true);
          }).catch(reject);
        });
      }).catch(reject);
    });
  }
}

module.exports = {
  key: 'add',
  synopsis: 'add a service to an app',
  usage: 'arigato add <service-name>',
  Command: AddService,
  example: `\tlocalhost$ arigato add sendgrid`
};
