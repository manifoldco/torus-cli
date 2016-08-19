'use strict';

var Promise = require('es6-promise').Promise;

var validate = require('../validate');
var output = require('../cli/output');

var validator = validate.build({
  org: validate.slug,
  email: validate.email
});

var approve = exports;

approve.output = {};
approve.output.success = output.create(function (opts) {
  var email = opts.invite.body.email;
  console.log('You have approved ' + email + '\'s invitation.');
  console.log();
  console.log('They are now a member of the organization!');
});

approve.output.failure = output.create(function () {
  console.log('Invitation approval failed; please try again.');
});

approve.execute = function (ctx) {
  return new Promise(function (resolve, reject) {
    var data = {
      org: ctx.option('org').value,
      email: ctx.params[0]
    };

    if (!data.org) {
      throw new Error('--org is required.');
    }

    if (!data.email) {
      throw new Error('You must supply an email as a parameter');
    }

    var errors = validator(data);
    if (errors.length > 0) {
      return reject(errors[0]);
    }

    return ctx.api.orgs.get({ name: data.org }).then(function (orgs) {
      var org = orgs[0];
      if (!org) {
        throw new Error('org not found: ' + data.org);
      }

      var opts = {
        org_id: org.id,
        email: data.email,
        state: 'accepted'
      };
      return ctx.api.invites.list(opts).then(function (invites) {
        if (invites.length !== 1) {
          throw new Error(
            'Cannot find an accepted invitation for ' + data.email);
        }

        var invite = invites[0];

        return ctx.api.invites.approve({}, { id: invite.id }, function (event) {
          console.log(event.message);
        })
        .then(function () {
          return resolve({
            org: org,
            invite: invite
          });
        });
      });
    }).catch(reject);
  });
};
