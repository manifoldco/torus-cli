'use strict';

const log = require('../util/log').get('cmds/init');
const CommandInterface = require('../command').CommandInterface;

const errors = require('../errors');
const users = require('../users');
const apps = require('../apps');
const ARIGATO_FILE_NAME = require('../descriptors/arigato').FILE_NAME;

class Init extends CommandInterface {

  execute (argv) {
    return new Promise((resolve, reject) => {
      var appName = argv._[0];
      users.loggedIn().then((loggedIn) => {
        if (!loggedIn) {
          log.error('You must be logged in to initialize an app.');
          return resolve(false);
        }

        return apps.init(appName);
      }).then((results) => {
        log.print('Your application '+results.app.name+' has been created');
        log.print('We\'ve created a unique environment for you as well');
        log.print(
          'An '+ARIGATO_FILE_NAME+' file has been added to your project root');
        resolve(true);
      }).catch((err) => {
        if (err instanceof errors.ValidationError) {
          log.error('Invalid app name provided: '+err.message);
          return resolve(false);
        }

        reject(err);
      });
    });
  }
}

module.exports = {
  key: 'init',
  synopsis: 'initialzie an arigato app in the root of your project',
  usage: 'arigato init <my-app>',
  Command: Init,
  example: `\tlocalhost$ arigato init api`
};
