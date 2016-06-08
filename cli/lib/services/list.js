'use strict';

var servicesList = exports;
var Promise = require('es6-promise').Promise;
var _ = require('lodash');

var Session = require('../session');
var output = require('../cli/output');
var client = require('../api/client').create();
var validate = require('../validate');

servicesList.output = {};

servicesList.output.success = output.create(function (services) {
  var length = services.length;

  var msg = services.map(function (service) {
    return _.padStart(service.body.name, service.body.name.length + 1);
  }).join('\n');

  console.log(' total (' + length + ')\n ---------\n' + msg);
});

servicesList.output.failure = output.create(function () {
  console.log('Retrieval of services failed!');
});

/**
 * List services
 *
 * @param {object} ctx
 */
servicesList.execute = function (ctx) {
  return new Promise(function (resolve, reject) {
    if (!(ctx.session instanceof Session)) {
      return reject(new TypeError('Session object not on Context'));
    }

    var orgName = ctx.options && ctx.options.org && ctx.options.org.value;

    if (!orgName) {
      return reject(new Error('--org is (temporarily) required.'));
    }

    var validOrgName = validate.slug(orgName);

    if (!_.isBoolean(validOrgName)) {
      return reject(new Error(validOrgName));
    }

    client.auth(ctx.session.token);

    return client.get({
      url: '/orgs',
      qs: { name: orgName }
    }).then(function (res) {
      var org = res.body && res.body[0];

      if (!_.isObject(org)) {
        return reject(new Error('The org could not be found'));
      }

      return client.get({
        url: '/services',
        qs: { org_id: org.id }
      }).then(resolve)
        .catch(reject);
    });
  });
};
