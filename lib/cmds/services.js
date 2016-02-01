'use strict';

const _ = require('lodash');

const log = require('../util/log').get('cmds/services');
const CommandInterface = require('../command').CommandInterface;
const ServiceRegistry = require('../descriptors/registry').Registry;
const Arigato = require('../descriptors/arigato').Arigato;

class ListServices extends CommandInterface {

  execute () {
    return new Promise((resolve, reject) => {
      var registry = new ServiceRegistry();

      Promise.all([
        registry.services(),
        Arigato.find(process.cwd())
      ]).then((results) => {
        var list = results[0];
        var installedServices = [];
        if (results[1] && _.isFunction(results[1].get)) {
          installedServices = results[1].get('services');
        }

        var intro = 'You can add the following services to your app'+
                    ' (* = already installed):';
        var howToAdd = 'Add a service to app using `arigato add <service>`';

        var listText = '';
        _.each(list, (service) => {
          var name = service.get('name');
          var installed = (installedServices.indexOf(name) > -1) ?
            '*' : '';
          listText += `    ${name}${installed}\n`;
        });

        log.print(`${intro}\n\n${listText}\n${howToAdd}`);
        resolve(true);
      }).catch(reject);
    });
  }
}

module.exports = {
  key: 'services',
  synopsis: 'list all installable services',
  usage: 'arigato services',
  Command: ListServices,
  example: `\tlocalhost$ arigato services`
};
