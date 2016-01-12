'use strict';

const inquirer = require('inquirer');

const users = require('../users');
const keys = require('../keys');
const log = require('../util/log').get('cmds/signup');
const CommandInterface = require('../command').CommandInterface;
const validation = require('../util/validation');

const QUESTIONS = [
  validation.fullName,
  validation.email,
  validation.password
];

class Signup extends CommandInterface {
  execute () {
    return new Promise((resolve, reject) => {
      // TODO: Come back and support ocnfirm password
      users.loggedIn().then((loggedIn) => {
        if (!loggedIn) {
          return Promise.resolve();
        }

        log.warn('You are currently logged in, logging you out!');
        return users.logout();
      }).then(() => {
        inquirer.prompt(QUESTIONS, (answers) => {
          var params = {
            name: answers.full_name,
            email: answers.email,
            password: answers.password,
          };

          users.create(params).then((user) => {
            var msg = `Success! We have created your account ${user.name}.`;
            log.print(msg);
            return users.login(params);
          }).then((results) => {
            var opts = {
              email:  params.email,
              password: params.password,
              uuid: results.user.uuid,
              session_token: results.session_token
            };

            return keys.create(opts);
          }).then(() => {
            log.print('Your pgp key has been generated!');
            resolve(true);
          }).catch(reject);
        });
      });
    });
  }
}

module.exports = {
  key: 'signup',
  synopsis: 'registers a user account',
  usage: 'arigato signup',
  Command: Signup,
  example: `\tlocalhost$ arigato signup`
};
