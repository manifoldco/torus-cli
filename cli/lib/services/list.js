'use strict';

var servicesList = exports;
var Promise = require('es6-promise').Promise;
var _ = require('lodash');

var Session = require('../session');
var output = require('../cli/output');
var client = require('../api/client').create();
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

    var orgName = ctx.option('org').value;
    var projectName = ctx.option('project').value;

    if (!orgName) {
      return reject(new Error('--org is required.'));
    }

    var errors = validator({
      org: orgName,
      project: projectName
    });
    if (errors.length > 0) {
      return reject(errors[0]);
    }

    client.auth(ctx.session.token);

    return client.get({
      url: '/orgs',
      qs: { name: orgName }
    }).then(function (res) {
      var org = res.body && res.body[0];

      if (!_.isObject(org)) {
        return reject(new Error('org not found: ' + orgName));
      }

      return Promise.all([
        client.get({
          url: '/projects',
          qs: {
            org_id: org.id
          }
        }),
        client.get({
          url: '/services',
          qs: {
            org_id: org.id
          }
        })
      ]).then(function (results) {
        var projects = results[0] && results[0].body;
        var services = results[1] && results[1].body;

        if (projectName) {
          projects = projects.filter(function (project) {
            return (project.body.name === projectName);
          });

          if (projects.length === 0) {
            return reject(new Error('project not found: ' + projectName));
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
