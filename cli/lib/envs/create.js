'use strict';

var envCreate = exports;

var Promise = require('es6-promise').Promise;
var _ = require('lodash');

var output = require('../cli/output');
var validate = require('../validate');
var Prompt = require('../cli/prompt');

var client = require('../api/client').create();

envCreate.output = {};

envCreate.output.success = output.create(function () {
  console.log('Environment created');
});

envCreate.output.failure = output.create(function () {
  console.log('Environment creation failed, please try again');
});

var validator = validate.build({
  name: validate.slug,
  project: validate.slug
});

/**
 * Create prompt for env create
 *
 * @param {object} ctx - Command context
 */
envCreate.execute = function (ctx) {
  return new Promise(function (resolve, reject) {
    var data = {
      name: ctx.params[0],
      org: ctx.option('org').value,
      project: ctx.option('project').value
    };

    if (!data.org) {
      throw new Error('--org is required.');
    }

    client.auth(ctx.session.token);

    var getOrgs = {
      url: '/orgs',
      qs: { name: data.org }
    };
    client.get(getOrgs).then(function (orgResult) {
      var org = orgResult.body && orgResult.body[0];
      if (!org) {
        throw new Error('org not found: ' + data.org);
      }

      return org;
    }).then(function (org) {
      var getProjects = {
        url: '/projects',
        qs: {
          org_id: org.id
        }
      };

      if (data.project) {
        getProjects.qs.name = data.project;
      }

      return client.get(getProjects).then(function (projectResult) {
        var projects = projectResult.body;

        if (!Array.isArray(projects)) {
          throw new Error('Invalid result from project retrieval');
        }

        if (data.project && projects.length !== 1) {
          throw new Error('project not found: ' + data.project);
        }

        if (projects.length === 0) {
          throw new Error(
            'You must create a project before creating an environment');
        }

        if (data.name && data.project) {
          var errors = validator(data);
          if (errors.length > 0) {
            return reject(errors[0]);
          }

          return envCreate._execute(org, projects, data).then(resolve);
        }

        var projectNames = _.map(projects, 'body.name');
        return envCreate._prompt(data, projectNames).then(function (input) {
          return envCreate._execute(org, projects, input).then(resolve);
        });
      });
    })
    .catch(reject);
  });
};

/**
 * Create prompt promise
 */
envCreate._prompt = function (defaults, projectNames) {
  var prompt = new Prompt({
    stages: envCreate._questions,
    defaults: defaults,
    questionArgs: [
      projectNames
    ]
  });

  return prompt.start().then(function (answers) {
    return _.omitBy(_.extend({}, defaults, answers), _.isUndefined);
  });
};

envCreate._execute = function (org, projects, input) {
  return new Promise(function (resolve, reject) {
    var project = _.find(projects, function (p) {
      return (p.body.name === input.project &&
              p.body.org_id === org.id);
    });

    if (!project) {
      throw new Error('project not found: ' + input.project);
    }

    var postEnvs = {
      url: '/envs',
      json: {
        body: {
          org_id: org.id,
          project_id: project.id,
          name: input.name
        }
      }
    };

    return client.post(postEnvs).then(function (result) {
      var env = result && result.body[0];
      if (!env) {
        throw new Error('Invalid service creation result');
      }

      resolve({
        project: project,
        env: env
      });
    }).catch(reject);
  });
};

envCreate._questions = function (projectNames) {
  return [
    [
      {
        name: 'name',
        message: 'Environment name',
        validate: validate.slug
      },
      {
        type: 'list',
        name: 'project',
        message: 'Project',
        choices: projectNames
      }
    ]
  ];
};

