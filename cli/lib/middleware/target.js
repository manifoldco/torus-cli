'use strict';

var Promise = require('es6-promise').Promise;

var targetMap = require('../context/map');
var Target = require('../context/target');
var client = require('../api/client').create();

module.exports = function () {
  return function (ctx) {
    return new Promise(function (resolve, reject) {
      targetMap.get().then(function (result) {
        ctx.target = new Target(result);

        if (ctx.session && !ctx.target.environment) {
          client.auth(ctx.session.token);

          return client.get({
            url: '/users/self'
          }).then(function (res) {
            var user = res.body && res.body[0];
            if (!user) {
              return reject(new Error('Could not find the user'));
            }

            ctx.target.flags({
              environment: 'dev-' + user.body.username
            });

            return resolve();
          });
        }

        return resolve();
      }).catch(reject);
    });
  };
};
