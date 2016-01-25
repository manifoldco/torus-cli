'use strict';

const log = require('../util/log').get('cmds/add');
const CommandInterface = require('../command').CommandInterface;
const ServiceRegistry = require('../descriptors/registry').Registry;
const questions = require('../questions');

const errors = require('../errors');
const users = require('../users');
const services = require('../services');

class AddService extends CommandInterface {

  _prepare (argv) {
    return new Promise((resolve, reject) => {
      if (argv._.length < 1) {
        return reject(new errors.ValidationError(
          'You must supply a service to add'
        ));
      }

      var name = argv._[0].toLowerCase();
      var registry = new ServiceRegistry();
      users.loggedIn().then((loggedIn) => {
        if (!loggedIn) {
          log.error('You must be logged in to add a service to an app.');
          return resolve({ success: false });
        }

        return registry.get(name);
      }).then((deployment) => {
        if (!deployment) {
          log.error('Service does not exist: '+argv._[0]);
          return resolve({ success: false });
        }

        return deployment.descriptor();
      }).then((descriptor) => {
        resolve({
          success: true,
          descriptor: descriptor,
          name: name,
          registry: registry
        });
      }).catch(reject);   
    });
  }

  execute (argv) {
    return new Promise((resolve, reject) => {
      this._prepare(argv).then((results) => {
        if (!results.success) {
          return resolve(false);
        }

        return questions.prompt(results.descriptor).then((answers) => {
          var opts = {
            name: results.name,
            descriptor: results.descriptor,
            provided: answers
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
