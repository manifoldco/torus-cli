'use strict';

var list = exports;
var Promise = require('es6-promise').Promise;
var _ = require('lodash');

var output = require('../cli/output');
var validate = require('../validate');

list.output = {};

list.output.success = output.create(function (payload) {
  var policy = payload.policy.body.policy;
  var team = payload.team.body;

  console.log(policy.name + ' has been detached from the ' + team.name + ' team.');
});

list.output.failure = output.create(function () {
  console.log('Policy could not be detached, please try again!');
});

var validator = validate.build({
  org: validate.slug,
  policy: validate.slug,
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

    if (ctx.params.length < 1) {
      return reject(new Error('policy [name] and [team] are required'));
    }

    var data = {
      org: ctx.target.org,
      policy: ctx.params[0],
      team: ctx.params[1]
    };

    if (!data.org) {
      return reject(new Error('--org is required.'));
    }

    var errors = validator(data);
    if (errors.length > 0) {
      return reject(errors[0]);
    }

    var payload = {};
    return ctx.api.orgs.get({ name: data.org }).then(function (res) {
      payload.org = res[0];

      if (!_.isObject(payload.org)) {
        return reject(new Error('org not found: ' + data.org));
      }

      return Promise.all([
        ctx.api.policies.get({
          org_id: payload.org.id,
          name: data.policy
        }),
        ctx.api.teams.get({
          org_id: payload.org.id,
          name: data.team
        })
      ]);
    }).then(function (results) {
      payload.policy = results[0] && results[0][0];
      payload.team = results[1] && results[1][0];

      if (!payload.policy) {
        throw new Error('policy not found: ' + data.policy);
      }

      if (!payload.team) {
        throw new Error('team not found: ' + data.team);
      }

      return ctx.api.policyAttachments.get({
        org_id: payload.org.id,
        owner_id: payload.team.id,
        policy_id: payload.policy.id
      });
    })
    .then(function (results) {
      payload.policyAttachment = results[0];

      if (!payload.policyAttachment) {
        throw new Error('policy does not appear to be attached');
      }

      return ctx.api.policyAttachments.delete({
        id: payload.policyAttachment.id
      });
    })
    .then(function () {
      resolve(payload);
    })
    .catch(reject);
  });
};
