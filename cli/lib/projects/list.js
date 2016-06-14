'use strict';

var _ = require('lodash');
var Promise = require('es6-promise').Promise;

var output = require('../cli/output');
var validate = require('../validate');
var client = require('../api/client').create();

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
    var projects = projectsByOrg[org.id];

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
    client.auth(ctx.session.token);

    var orgName = ctx.options.org.value;
    if (orgName) {
      var errors = validator({ org: orgName });
      if (errors.length > 0) {
        return reject(errors[0]);
      }
    }

    // if orgname given then just get that specific org data
    // otherwise get everything.
    var getOrgs = {
      url: '/orgs',
      qs: {}
    };

    if (orgName) {
      getOrgs.qs.name = orgName;
    }

    return client.get(getOrgs).then(function (results) {
      if (!results.body || results.body.length === 0) {
        if (orgName) {
          return reject(new Error('Could not find org: ' + orgName));
        }

        return reject(new Error('Could not find any orgs'));
      }

      // If we're selecting a sepcific org then just filtering it
      // out of the results and return a single array.
      if (orgName) {
        var org = _.find(results.body, function (o) {
          return o.body.name === orgName;
        });

        return [org];
      }

      return results.body;
    }).then(function (orgs) {
      var getProjects = {
        url: '/projects',
        qs: {
          org_id: orgs.map(function (org) {
            return org.id;
          })
        }
      };

      return client.get(getProjects).then(function (results) {
        if (!results.body) {
          return reject(new Error('Could not find project(s'));
        }

        return resolve({
          orgs: orgs,
          projects: results.body
        });
      });
    })
    .catch(reject);
  });
};
