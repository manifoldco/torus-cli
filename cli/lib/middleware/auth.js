'use strict';

var Promise = require('es6-promise').Promise;

module.exports = function () {
  return function (ctx) {
    return new Promise(function (resolve, reject) {
      var name = ctx.program.name;
      var slug = ctx.slug;
      ctx.api.session.get().then(function () {
        return resolve();
      }).catch(function () {
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
