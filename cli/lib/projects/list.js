'use strict';

var _ = require('lodash');
var Promise = require('es6-promise').Promise;

var output = require('../cli/output');
var validate = require('../validate');

var list = exports;
list.output = {};

var validator = validate.build({
  org: validate.slug
});

list.output.success = output.create(function (payload) {
  var orgIdMap = {};
  payload.orgs.forEach(function (org) {
    orgIdMap[org.id] = org;
  });

  var numOrgs = Object.keys(orgIdMap).length;
  var projectsByOrg = _.groupBy(payload.projects, 'body.org_id');

  _.each(orgIdMap, function (org, i) {
    var projects = projectsByOrg[org.id] || [];

    var msg = ' ' + org.body.name + ' org (' + projects.length + ')\n';
    msg += ' ' + new Array(msg.length - 1).join('-') + '\n'; // underline

    msg += projects.map(function (project) {
      return _.padStart(project.body.name, project.body.name.length + 1);
    }).join('\n');

    if (i + 1 !== numOrgs) {
      msg += '\n';
    }

    console.log(msg);
  });
});

list.output.failure = output.create(function () {
  console.log('Project retrieval failed, please try again.');
});

list.execute = function (ctx) {
  return new Promise(function (resolve, reject) {
    ctx.target.flags({
      org: ctx.option('org').value
    });

    var orgName = ctx.target.org;
    if (orgName) {
      var errors = validator({ org: orgName });
      if (errors.length > 0) {
        return reject(errors[0]);
      }
    }

    // if orgname given then just get that specific org data
    // otherwise get everything.
    var qs = {};
    if (orgName) {
      qs.name = orgName;
    }

    return ctx.api.orgs.get(qs).then(function (orgs) {
      if (orgs.length === 0) {
        if (orgName) {
          return reject(new Error('Could not find org: ' + orgName));
        }

        return reject(new Error('Could not find any orgs'));
      }

      // If we're selecting a sepcific org then just filtering it
      // out of the orgs and return a single array.
      if (orgName) {
        var org = _.find(orgs, function (o) {
          return o.body.name === orgName;
        });

        return [org];
      }

      return orgs;
    }).then(function (orgs) {
      var orgIds = orgs.map(function (org) { return org.id; });
      var orgsQs = { org_id: orgIds };
      return ctx.api.projects.get(orgsQs).then(function (projects) {
        if (!projects) {
          return reject(new Error('Could not find project(s)'));
        }

        return resolve({
          orgs: orgs,
          projects: projects
        });
      });
    })
    .catch(reject);
  });
};
