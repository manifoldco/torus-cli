'use strict';

var Promise = require('es6-promise').Promise;
var errors = require('common/errors');

var series = require('../cli/series');
var validate = require('../validate');
var output = require('../cli/output');

var generate = exports;

var validator = validate.build({
  org: validate.slug
});

generate.output = {};

generate.output.failure = output.create(function () {
  console.log('Keypair generation failed; please try again!');
});

generate.execute = function (ctx) {
  if (ctx.option('all').value) {
    return generate._all(ctx);
  }

  return generate._all(ctx);
};

/**
 * Command for re-generating keypairs for a single organization.
 */
generate._single = function (ctx) {
  return new Promise(function (resolve, reject) {
    ctx.target.flags({
      org: ctx.option('org').value
    });

    var data = { org: ctx.target.org };
    if (!data.org) {
      throw new errors.BadRequest('--org is required.');
    }

    var errs = validator(data);
    if (errs.length > 0) {
      return reject(errs[0]);
    }

    return ctx.api.orgs.get({ name: data.org }).then(function (orgs) {
      var org = orgs[0];
      if (!org) {
        throw new errors.NotFound('org not found: ' + data.org);
      }

      console.log(
        'Generating signing and encryption keypairs for', data.org, '\n');
      return ctx.api.keypairs.generate({ org_id: org.id })
      .then(function () {
        console.log('Keypair generation successful.\n');
      });
    }).catch(reject);
  });
};

/**
 * Command for generating keypairs for all orgs missing valid existing keys
 */
generate._all = function (ctx) {
  return Promise.all([
    ctx.api.orgs.get(),
    ctx.api.keypairs.list()
  ]).then(function (results) {
    var orgs = results[0];
    var keypairs = results[1];

    var hasKeys = {};
    keypairs.forEach(function (keypair) {
      var orgId = keypair.public_key.body.org_id;
      if (!hasKeys[orgId]) {
        hasKeys[orgId] = {};
      }

      hasKeys[orgId][keypair.public_key.body.type] = true;
    });

    var toCreate = orgs.map(function (o) {
      if (hasKeys[o.id] && hasKeys[o.id].signing && hasKeys[o.id].encryption) {
        return undefined;
      }

      return function () {
        console.log(
          'Generating signing and encryption keypairs for', o.body.name, '\n');

        return ctx.api.keypairs.generate({ org_id: o.id });
      };
    }).filter(function (v) {
      return v !== undefined;
    });

    if (toCreate.length === 0) {
      console.log(
        'All organizations have valid signing and encryption keys.\n');
      return Promise.resolve();
    }

    return series(toCreate).then(function () {
      console.log('Keypair generation complete.\n');
    });
  });
};
