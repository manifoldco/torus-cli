'use strict';

var create = exports;

var _ = require('lodash');
var Promise = require('es6-promise').Promise;

var output = require('../cli/output');
var validate = require('../validate');
var Prompt = require('../cli/prompt');
var client = require('../api/client').create();

var USER_TYPE = 'user';

create.output = {};

var validator = validate.build({
  name: validate.slug,
  org: validate.slug
});

create.output.success = output.create(function () {
  console.log('Team created.');
});

create.output.failure = output.create(function () {
  console.log('Team creation failed, please try again');
});

create.execute = function (ctx) {
  return new Promise(function (resolve, reject) {
    client.auth(ctx.session.token);

    var orgs;
    client.get({ url: '/orgs' }).then(function (results) {
      if (!results.body || results.body.length < 1) {
        throw new Error('Could not locate organizations');
      }

      orgs = results.body;

      ctx.target.flags({
        org: ctx.option('org').value
      });

      var data = {
        name: ctx.params[0],
        org: ctx.target.org
      };

      if (data.name && data.org) {
        var errors = validator(data);
        if (errors.length > 0) {
          return reject(errors[0]);
        }

        return data;
      }

      return create._prompt(data, _.map(orgs, 'body.name'));
    }).then(function (data) {
      var org = _.find(orgs, function (o) {
        return o.body.name === data.org;
      });

      if (!org) {
        throw new Error('unknown org: ' + data.org);
      }

      return client.post({
        url: '/teams',
        json: {
          body: {
            org_id: org.id,
            name: data.name,
            type: USER_TYPE
          }
        }
      });
    })
    .then(function (result) {
      resolve(result.body);
    })
    .catch(reject);
  });
};

create._prompt = function (defaults, orgNames) {
  var prompt = new Prompt({
    stages: create._questions,
    defaults: defaults,
    questionArgs: [
      orgNames
    ]
  });

  return prompt.start().then(function (answers) {
    return _.omitBy(_.extend({}, defaults, answers), _.isUndefined);
  });
};

create._questions = function (orgNames) {
  return [
    [
      {
        name: 'name',
        message: 'Team Name',
        validate: validate.slug
      },
      {
        type: 'list',
        name: 'org',
        message: 'organization the team belongs to',
        choices: orgNames
      }
    ]
  ];
};
