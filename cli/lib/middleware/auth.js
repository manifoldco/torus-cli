'use strict';

var Promise = require('es6-promise').Promise;

var login = require('../login/index');

module.exports = function () {
  return function (ctx) {
    return new Promise(function (resolve, reject) {
      var name = ctx.program.name;
      var slug = ctx.slug;
      ctx.api.session.get().then(function () {
        ctx.loggedIn = true;
        return resolve();
      }).catch(function (err) {
        if (err.type !== 'unauthorized') {
          return reject(err);
        }

        // if the email and password are set in the environment; then we can
        // try and login with them!
        if (process.env.AG_EMAIL && process.env.AG_PASSWORD) {
          var inputs = {
            email: process.env.AG_EMAIL,
            passphrase: process.env.AG_PASSWORD
          };

          console.log('Attempting to login with email:', inputs.email);
          return login.subcommand(ctx, inputs).then(function () {
            console.log('Logged in successfully!\n');

            ctx.loggedIn = true;
            return resolve();
          }).catch(reject);
        }

        console.log('You must be logged-in to execute \'' +
                    name + ' ' + slug + '\'');
        console.log();
        console.log(
          'Login using \'' + name + ' login\' or create an account with ' +
          '\'' + name + ' signup\'');
        return reject();
      });
    });
  };
};
