'use strict';

var orgsList = exports;

var Promise = require('es6-promise').Promise;
var _ = require('lodash');

var output = require('../cli/output');

orgsList.output = {};

/**
 * Success
 *
 * @param {object} payload - Response object
 */
orgsList.output.success = output.create(function (payload) {
  var orgs = payload.orgs;
  var self = payload.self;

  var msg = ' orgs (' + orgs.length + ')\n';

  var personalOrg = _.takeWhile(orgs, function (o) {
    return o.body.name === self.body.username;
  });

  msg += ' ' + new Array(msg.length - 1).join('-') + '\n'; // underline
  msg += ' ' + personalOrg[0].body.name + ' [personal]\n'; // personal org first

  msg += _.difference(orgs, personalOrg).map(function (org) {
    return _.padStart(org.body.name, org.body.name.length + 1);
  }).join('\n');

  console.log(msg);
});

orgsList.output.failure = output.create(function () {
  console.log('Retrieval of orgs failed!');
});

/**
 * List orgs
 *
 * @param {object} ctx - Command context
 */
orgsList.execute = function (ctx) {
  return new Promise(function (resolve, reject) {
    return Promise.all([
      ctx.api.users.self(),
      ctx.api.orgs.get()
    ]).then(function (res) {
      var self = res[0][0];
      var orgs = res[1];

      if (!_.isArray(orgs) || !_.isObject(self)) {
        return reject(new Error('No orgs found'));
      }

      return resolve({
        orgs: orgs,
        self: self
      });
    }).catch(reject);
  });
};
