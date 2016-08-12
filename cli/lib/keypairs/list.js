'use strict';

var Promise = require('es6-promise').Promise;
var _ = require('lodash');
var Table = require('cli-table2');
var errors = require('common/errors');

var validate = require('../validate');
var output = require('../cli/output');

var keypairsList = exports;

var validator = validate.build({
  org: validate.slug
}, false);

keypairsList.output = {};

keypairsList.output.success = output.create(function (ctx, payload) {
  if (ctx.option('org').value) {
    console.log(
      'Listing your keys for the ' + ctx.option('org').value + ' org.\n');
  }

  var orgIndex = _.keyBy(payload.orgs, 'id');
  var table = new Table({
    head: [
      'ID',
      'Org',
      'Key Type',
      'Valid',
      'Creation Date'
    ]
  });

  var hasKeys = {};
  var rows = payload.keypairs.map(function (keypair) {
    // You can have keys for orgs you no longer belong too.
    var org = orgIndex[keypair.public_key.body.org_id];
    if (!org) {
      return [];
    }

    if (!hasKeys[org.id]) {
      hasKeys[org.id] = {};
    }
    hasKeys[org.id][keypair.public_key.body.type] = true;

    return [
      keypair.public_key.id,
      org.body.name,
      keypair.public_key.body.type,
      'Yes', // all keys are valid today (you can't revoke)
      new Date(keypair.public_key.body.created_at).toISOString()
    ];
  }).filter(function (row) {
    return row.length !== 0;
  });

  if (rows.length === 0) {
    return console.log('No keypairs found.');
  }

  table.push.apply(table, rows);
  console.log(table.toString());

  var noMissingKeys = payload.orgs.every(function (org) {
    if (!hasKeys[org.id]) {
      return false;
    }

    return hasKeys[org.id].signing === true &&
      hasKeys[org.id].encryption === true;
  });

  if (!noMissingKeys) {
    console.log('\nYou are missing keys for an organization.\n');
    console.log(
      'You can generate missing keys using \'ag keypairs:generate --all\'');
  }

  return Promise.resolve();
});

keypairsList.output.failure = output.create(function () {
  console.log('Retrieval of keypairs failed!');
});

keypairsList.execute = function (ctx) {
  return new Promise(function (resolve, reject) {
    ctx.target.flags({
      // If 'all' is set (via -a or --all) then ignore the context!
      org: (ctx.option('all').value) ? null : ctx.option('org').value
    });

    var data = { org: ctx.target.org };
    if (data.org) {
      var errs = validator(data);
      if (errs.length > 0) {
        return reject(errs[0]);
      }
    }

    // If an org is provided only show those keypairs, otherwise, get back all
    // of the orgs for the user.
    var opts = (data.org) ? { name: data.org } : {};
    return ctx.api.orgs.get(opts).then(function (orgs) {
      if (data.org && orgs.length < 1) {
        throw new errors.NotFound('org not found: ' + data.org);
      }

      var getOpts = (data.org) ? { org_id: orgs[0].id } : {};
      return ctx.api.keypairs.list(getOpts).then(function (pairs) {
        resolve({
          orgs: orgs,
          keypairs: pairs
        });
      });
    }).catch(reject);
  });
};
