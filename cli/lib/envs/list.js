'use strict';

var envsList = exports;

var Promise = require('es6-promise').Promise;
var _ = require('lodash');

var validate = require('../validate');
var output = require('../cli/output');
var client = require('../api/client').create();
var Session = require('../session');

envsList.output = {};

/**
 * Success
 *
 * @param {object} payload - Response object
 */
envsList.output.success = output.create(function (payload) {
  var envs = payload.envs;
  var projects = payload.projects;

  var envsByProject = _.groupBy(envs, 'body.project_id');

  _.each(projects, function (project) {
    var projectEnvs = envsByProject[project.id] || [];
    var projectName = project.body.name;

    var msg = ' ' + projectName + ' projects (' + projectEnvs.length + ')\n';

    msg += ' ' + new Array(msg.length - 1).join('-') + '\n'; // underline
    msg += projectEnvs.map(function (env) {
      return _.padStart(env.body.name, env.body.name.length + 1);
    }).join('\n');

    if (envs !== _.findLast(projectEnvs)) {
      msg += '\n';
    }

    console.log(msg);
  });
});

envsList.output.failure = output.create(function () {
  console.log('Retrieval of envs failed!');
});

var validator = validate.build({
  org: validate.slug,
  project: validate.slug
}, false);

/**
 * List envs
 *
 * @param {object} ctx - Command context
 */
envsList.execute = function (ctx) {
  return new Promise(function (resolve, reject) {
    if (!(ctx.session instanceof Session)) {
      throw new TypeError('Session object missing on Context');
    }

    ctx.target.flags({
      org: ctx.option('org').value
    });

    var data = {
      org: ctx.target.org,
      project: ctx.option('project').value
    };

    if (!data.org) {
      throw new Error('--org is required.');
    }

    var errors = validator(data);
    if (errors.length > 0) {
      return reject(errors[0]);
    }

    client.auth(ctx.session.token);

    return client.get({
      url: '/orgs',
      qs: { name: data.org }
    }).then(function (res) {
      var org = res.body && res.body[0];

      if (!_.isObject(org)) {
        return reject(new Error('org not found: ' + data.org));
      }

      // XXX: This returns all envs and all projects for an org, over time,
      // as the number of projects and environments scale in an org this will
      // fall over and get really slow.
      return Promise.all([
        client.get({
          url: '/projects',
          qs: {
            org_id: org.id
          }
        }),
        client.get({
          url: '/envs',
          qs: {
            org_id: org.id
          }
        })
      ]).then(function (results) {
        var projects = results[0] && results[0].body;
        var envs = results[1] && results[1].body;

        if (data.project) {
          projects = projects.filter(function (project) {
            return (project.body.name === data.project);
          });

          if (projects.length === 0) {
            throw new Error('project not found: ' + data.project);
          }

          envs = envs.filter(function (env) {
            return (projects[0].id === env.body.project_id);
          });
        }

        return resolve({
          projects: projects || [],
          envs: envs || []
        });
      });
    }).catch(reject);
  });
};
