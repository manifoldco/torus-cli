/* eslint-disable */

'use strict';

// TODO: Accept and validate teams
// TODO: Import the policy to an organization
// TODO: Attach the policy to a team
// TODO: Dream up some naming conventions for policies

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
  console.log('Boo-hoo, policies')
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

    // harvest(ctx);

    var params = {
      actions: ctx.params[0],
      path: ctx.params[1],
      team: ctx.params[2]
    };

    var pathSegments = ctx.params[1].split('/');

    var secret = _.takeRight(pathSegments);
    var path = _.take(pathSegments, pathSegments.length - 1).join('/');

    if (!resources.validPath(path, secret)) {
      reject(new Error('boom'));
    }

    var resourcesMap = resources.fromPath(path, secret);
    var extendedResources = resources.explode(resourcesMap);

    var policy = new Policy('generated policy');

    var defaultActions = [Statement.ACTIONS.READ, Statement.ACTIONS.LIST]

    _.each(extendedResources, function (r) {
      var statement = new Statement(EFFECT_ALLOW);
      var actions = r.split('/').length < 6 ? DEFAULT_ACTIONS : params.actions;

      statement.setActions(actions);
      statement.setResource(r);

      policy.addStatement(statement);
    });

    var org;
    var team;
    return ctx.api.orgs.get({ name: 'knotty-buoy' })
      .then(function (res) {
        org = res[0];

        if (!org) {
          return reject(new Error('Unknown org: ' + 'ddf'));
        }

        return ctx.api.teams.get({ name: params.team, org_id: org.id });
      })
      .then(function (res) {
        team = res[0];

        if (!team) {
          return reject(new Error('Unknown team: ' + params.team));
        }

        // return ctx.api.policies.create(policy)
      })
      .catch(reject);
  });
};
