'use strict';

var servicesList = exports;
var Promise = require('es6-promise').Promise;
var _ = require('lodash');

var Session = require('../session');
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
    if (!(ctx.session instanceof Session)) {
      return reject(new TypeError('Session object not on Context'));
    }

    ctx.target.flags({
      org: ctx.option('org').value
    });

    var data = {
      org: ctx.target.org,
      project: ctx.option('project').value
    };

    if (!data.org) {
      return reject(new Error('--org is required.'));
    }

    var errors = validator(data);
    if (errors.length > 0) {
      return reject(errors[0]);
    }

    return ctx.api.orgs.get({ name: data.org }).then(function (res) {
      var org = res[0];
      if (!_.isObject(org)) {
        return reject(new Error('org not found: ' + data.org));
      }

      // XXX: This returns all services and all projects for an org, over time,
      // as the number of projects and services scale in an org this will fall
      // over and get really slow.
      return Promise.all([
        ctx.api.projects.get({ org_id: org.id }),
        ctx.api.services.get({ org_id: org.id })
      ]).then(function (results) {
        var projects = results[0];
        var services = results[1];

        if (data.project) {
          projects = projects.filter(function (project) {
            return (project.body.name === data.project);
          });

          if (projects.length === 0) {
            return reject(new Error('project not found: ' + data.project));
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
