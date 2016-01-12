'use strict';

const _ = require('lodash');

const log = require('../util/log').get('cmds/services');
const CommandInterface = require('../command').CommandInterface;
const services = require('../services');

class ListServices extends CommandInterface {

  execute () {
    return new Promise((resolve, reject) => {
      services.supported().then((list) => {
        var intro = 'You can add the following services to your app:';
        var howToAdd = 'Add a service to app using `arigato add <service>`';

        var listText = '';
        _.each(list, (service) => {
          listText += `    ${service.name}\n`;
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
