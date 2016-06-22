'use strict';

var _ = require('lodash');
var Promise = require('es6-promise').Promise;

var output = require('../cli/output');
var validate = require('../validate');
var Prompt = require('../cli/prompt');
var client = require('../api/client').create();

var create = exports;
create.output = {};

var validator = validate.build({
  name: validate.slug,
  org: validate.slug
});

create.output.success = output.create(function () {
  console.log('Project created.');
});

create.output.failure = output.create(function () {
  console.log('Project creation failed, please try again');
});

create.execute = function (ctx) {
  return new Promise(function (resolve, reject) {
    client.auth(ctx.session.token);

    var getOrgs = {
      url: '/orgs'
    };

    var orgs;
    client.get(getOrgs).then(function (results) {
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
        throw new Error('Unknown org: ' + data.org);
      }

      var createProject = {
        url: '/projects',
        json: {
          body: {
            org_id: org.id,
            name: data.name
          }
        }
      };

      return client.post(createProject);
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
        message: 'Project Name',
        validate: validate.slug
      },
      {
        type: 'list',
        name: 'org',
        message: 'Organization the project belongs to',
        choices: orgNames
      }
    ]
  ];
};
