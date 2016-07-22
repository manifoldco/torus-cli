'use strict';

// TODO: Import the policy to an organization
// TODO: Attach the policy to a team
// TODO: Dream up some naming conventions for policies
// TODO: Pretty Output
// TODO: Remove return true on 88

var allow = exports;

var _ = require('lodash');
var Promise = require('es6-promise').Promise;
var Policy = require('./policy').Policy;
var Statement = require('./policy').Statement;
var output = require('../cli/output');
var resources = require('./resources');
var harvest = require('./harvest');

var EFFECT_ALLOW = Statement.EFFECTS.ALLOW;
var DEFAULT_ACTIONS = [Statement.ACTIONS.READ, Statement.ACTIONS.LIST];

allow.output = {};

allow.output.success = output.create(function () {
  console.log('Woo-hoo, policies');
});

allow.output.failure = output.create(function () {
  console.log('Boo-hoo, policies');
});

/**
 * Create prompt for allow
 *
 * @param {object} ctx - Command context
 */
allow.execute = function (ctx) {
  return new Promise(function (resolve, reject) {
    if (ctx.params.length < 2) {
      return reject(new Error('You must provide two parameters'));
    }

    var params = harvest(ctx);

    var pathSegments = params.path.split('/');

    var secret = _.takeRight(pathSegments);
    var path = _.take(pathSegments, pathSegments.length - 1).join('/');

    var resourceMap = resources.fromPath(path, secret);
    var extendedResources = resources.explode(resourceMap);

    var policy = new Policy('generated-policy');

    _.each(extendedResources, function (r) {
      var statement = new Statement(EFFECT_ALLOW);
      var actions = r.split('/').length < 6 ? DEFAULT_ACTIONS : params.actions;

      statement.setActions(actions);
      statement.setResource(r);

      policy.addStatement(statement);
    });

    var payload = {};
    return ctx.api.orgs.get({ name: pathSegments[1] })
      .then(function (res) {
        payload.org = res[0];

        if (!payload.org) {
          return reject(new Error('Unknown org: ' + pathSegments[1]));
        }

        return ctx.api.teams.get({ name: params.team, org_id: payload.org.id });
      })
      .then(function (res) {
        payload.team = res[0];

        if (!payload.team) {
          return reject(new Error('Unknown team: ' + params.team));
        }

        return ctx.api.policies.create({
          org_id: payload.org.id,
          policy: _.toPlainObject(policy)
        });
      })
      .then(function (res) {
        payload.policy = res;

        if (!payload.policy) {
          return reject(new Error('Error creating policy'));
        }

        return ctx.api.policyAttachments.create({
          org_id: payload.org.id,
          owner_id: payload.team.id,
          policy_id: payload.policy.id
        });
      })
      .then(function () { return payload; })
      .catch(reject);
  });
};
