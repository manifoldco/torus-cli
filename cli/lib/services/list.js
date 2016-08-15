'use strict';

var servicesList = exports;
var Promise = require('es6-promise').Promise;
var _ = require('lodash');

var output = require('../cli/output');
var validate = require('../validate');

servicesList.output = {};

servicesList.output.success = output.create(function (payload) {
  var projectIdMap = {};
  payload.projects.forEach(function (project) {
    projectIdMap[project.id] = project;
  });

  var numProjects = Object.keys(projectIdMap).length;
  var servicesByProject = _.groupBy(payload.services, 'body.project_id');

  _.each(projectIdMap, function (project, i) {
    var services = servicesByProject[project.id] || [];

    var msg = ' ' + project.body.name + ' project (' + services.length + ')\n';
    msg += ' ' + new Array(msg.length - 1).join('-') + '\n';

    msg += services.map(function (service) {
      return _.padStart(service.body.name, service.body.name.length + 1);
    }).join('\n');

    if (i + 1 !== numProjects) {
      msg += '\n';
    }

    console.log(msg);
  });
});

servicesList.output.failure = output.create(function () {
  console.log('Retrieval of services failed!');
});

var validator = validate.build({
  org: validate.slug,
  project: validate.slug
}, false);

/**
 * List services
 *
 * @param {object} ctx
 */
servicesList.execute = function (ctx) {
  return new Promise(function (resolve, reject) {
    ctx.target.flags({
      org: ctx.option('org').value,
      project: ctx.option('all').value ? null : ctx.option('project').value
    });

    var check = {
      org: ctx.target.org
    };

    if (!ctx.target.org) {
      return reject(new Error('--org is required.'));
    }

    if (ctx.option('all').value) {
      if (ctx.option('project').value) {
        return reject(new Error('project flag cannot be used with --all'));
      }
    } else {
      if (!ctx.target.project) {
        return reject(new Error('--project is required.'));
      }
      check.project = ctx.target.project;
    }

    var errors = validator(check);
    if (errors.length > 0) {
      return reject(errors[0]);
    }

    return ctx.api.orgs.get({ name: ctx.target.org }).then(function (res) {
      var org = res[0];
      if (!_.isObject(org)) {
        return reject(new Error('org not found: ' + ctx.target.org));
      }

      // XXX: This returns all services and all projects for an org, over time,
      // as the number of projects and services scale in an org this will fall
      // over and get really slow
      return Promise.all([
        ctx.api.projects.get({ org_id: org.id }),
        ctx.api.services.get({ org_id: org.id })
      ]).then(function (results) {
        var projects = results[0];
        var services = results[1];

        if (ctx.target.project) {
          projects = projects.filter(function (project) {
            return (project.body.name === ctx.target.project);
          });

          if (projects.length === 0) {
            return reject(new Error('project not found: ' + ctx.target.project));
          }

          services = services.filter(function (service) {
            return (projects[0].id === service.body.project_id);
          });
        }

        return resolve({
          projects: projects || [],
          services: services || []
        });
      });
    }).catch(reject);
  });
};
