'use strict';

var list = exports;
var Promise = require('es6-promise').Promise;
var _ = require('lodash');

var output = require('../cli/output');
var validate = require('../validate');

list.output = {};

list.output.success = output.create(function (payload) {
  var org = payload.org;
  var policies = payload.policies;
  var policyAttachments = payload.policyAttachments;
  var teams = payload.teams;

  var policyAttachmentsByPolicy = _.groupBy(policyAttachments, 'body.policy_id');
  var teamsById = _.reduce(teams, function (map, team) {
    map[team.id] = team;
    return map;
  }, {});

  var msg = ' Policies associated with the ' + org.body.name + ' org (' + policies.length + ')\n';
  msg += ' ' + new Array(msg.length - 1).join('-');

  _.each(policies, function (policy) {
    msg += ' \n ' + policy.body.policy.name;

    var attachments = policyAttachmentsByPolicy[policy.id] || [];

    if (attachments.length < 1) return;

    var teamNames = _.map(attachments, function (a) {
      return teamsById[a.body.owner_id].body.name;
    });

    msg += ' [' + teamNames.join(', ') + ']';
  });

  console.log(msg);
});

list.output.failure = output.create(function () {
  console.log('Retrieval of services failed!');
});

var validator = validate.build({
  org: validate.slug,
  team: validate.slug
}, false);

/**
 * List services
 *
 * @param {object} ctx
 */
list.execute = function (ctx) {
  return new Promise(function (resolve, reject) {
    ctx.target.flags({
      org: ctx.option('org').value
    });

    var data = {
      org: ctx.target.org,
      team: ctx.target.team
    };

    if (!data.org) {
      return reject(new Error('--org is required.'));
    }

    var errors = validator(data);
    if (errors.length > 0) {
      return reject(errors[0]);
    }

    return ctx.api.orgs.get({ name: data.org }).then(function (res) {
      var org = res[0];
      if (!_.isObject(org)) {
        return reject(new Error('org not found: ' + data.org));
      }

      return Promise.all([
        ctx.api.policies.get({ org_id: org.id }),
        ctx.api.policyAttachments.get({ org_id: org.id }),
        ctx.api.teams.get({ org_id: org.id })
      ]).then(function (results) {
        return resolve({
          org: org,
          policies: results[0] || [],
          policyAttachments: results[1] || [],
          teams: results[2] || []
        });
      });
    }).catch(reject);
  });
};
