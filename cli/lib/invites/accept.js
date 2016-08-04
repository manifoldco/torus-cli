'use strict';

var Promise = require('es6-promise').Promise;

var validate = require('../validate');
var output = require('../cli/output');

var validator = validate.build({
  org: validate.slug,
  email: validate.email,
  code: validate.inviteCode
});

var accept = exports;

accept.output = {};
accept.output.success = output.create(function () {
  console.log('You have accepted the invitation.');
  console.log();
  console.log('You will be added to the org once the administrator ' +
    ' has approved your invite.');
});

accept.output.failure = output.create(function () {
  console.log('Invitation acception failed; please try again.');
});

accept.validate = function (ctx) {
  return new Promise(function (resolve, reject) {
    var data = {
      org: ctx.option('org').value,
      email: ctx.params[0],
      code: ctx.params[1]
    };

    if (!data.org) {
      throw new Error('--org is required.');
    }

    if (!data.email) {
      throw new Error('You must supply an email as a parameter');
    }

    if (!data.code) {
      throw new Error('You must supply an invite code as a parameter');
    }

    var errors = validator(data);
    if (errors.length > 0) {
      return reject(errors[0]);
    }

    return resolve(data);
  });
};

accept.associate = function (ctx, params) {
  return ctx.api.invites.associate(params);
};

accept.finalize = function (ctx, params) {
  return ctx.api.invites.accept(params);
};
