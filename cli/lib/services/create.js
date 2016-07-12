'use strict';

var serviceCreate = exports;

var _ = require('lodash');
var Promise = require('es6-promise').Promise;

var output = require('../cli/output');
var validate = require('../validate');
var Prompt = require('../cli/prompt');

serviceCreate.output = {};

/**
 * Success output
 */
serviceCreate.output.success = output.create(function () {
  console.log('Service created.');
});

/**
 * Failure output
 */
serviceCreate.output.failure = output.create(function () {
  console.log('Service creation failed, please try again');
});

var validator = validate.build({
  name: validate.slug,
  project: validate.slug
});

/**
 * Create prompt for service create
 *
 * @param {object} ctx - Command context
 */
serviceCreate.execute = function (ctx) {
  return new Promise(function (resolve, reject) {
    ctx.target.flags({
      org: ctx.option('org').value,
      project: ctx.option('project').value
    });

    var data = {
      name: ctx.params[0],
      org: ctx.target.org,
      project: ctx.target.project
    };

    if (!data.org) {
      throw new Error('--org is required.');
    }


    return ctx.api.orgs.get({ name: data.org }).then(function (orgs) {
      var org = orgs[0];
      if (!org) {
        throw new Error('org not found: ' + data.org);
      }

      return org;
    }).then(function (org) {
      var qs = { org_id: org.id };
      if (data.project) {
        qs.name = data.project;
      }

      return ctx.api.projects.get(qs).then(function (projects) {
        if (!Array.isArray(projects)) {
          throw new Error('Invalid result from project retreival');
        }

        if (data.project && projects.length !== 1) {
          throw new Error('project not found: ' + data.project);
        }

        if (projects.length === 0) {
          throw new Error(
            'You must create a project before creating a service');
        }

        if (data.name && data.project) {
          var errors = validator(data);
          if (errors.length > 0) {
            return reject(errors[0]);
          }

          return serviceCreate._execute(ctx.api, org, projects, data)
            .then(resolve);
        }

        var projectNames = _.map(projects, 'body.name');
        return serviceCreate._prompt(data, projectNames).then(function (input) {
          return serviceCreate._execute(ctx.api, org, projects, input)
            .then(resolve);
        });
      });
    })
    .catch(reject);
  });
};

/**
 * Attempt to create service with supplied input
 *
 * @param {Object} org organization object
 * @param {Object} project array of possible projects
 * @param {Object} input user input
 */
serviceCreate._execute = function (api, org, projects, input) {
  return new Promise(function (resolve, reject) {
    var project = _.find(projects, function (p) {
      return (p.body.name === input.project &&
              p.body.org_id === org.id);
    });

    if (!project) {
      throw new Error('project not found: ' + input.project);
    }

    var data = {
      org_id: org.id,
      project_id: project.id,
      name: input.name
    };

    return api.services.create(data).then(function (result) {
      var service = result[0];
      if (!service) {
        throw new Error('Invalid service creation result');
      }

      resolve({
        project: project,
        service: service
      });
    }).catch(reject);
  });
};

/**
 * Create prompt promise
 */
serviceCreate._prompt = function (defaults, projectNames) {
  var prompt = new Prompt({
    stages: serviceCreate._questions,
    defaults: defaults,
    questionArgs: [
      projectNames
    ]
  });

  return prompt.start().then(function (answers) {
    return _.omitBy(_.extend({}, defaults, answers), _.isUndefined);
  });
};

/**
 * Service prompt questions
 */
serviceCreate._questions = function (projectNames) {
  return [
    [
      {
        name: 'name',
        message: 'Service name',
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
