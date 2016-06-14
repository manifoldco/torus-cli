'use strict';

var Promise = require('es6-promise').Promise;

var cpath = require('common/cpath');
var utils = require('common/utils');

var Session = require('../session');
var client = require('../api/client').create();
var cValue = require('./value');

var credentials = exports;

function getPath(user, project, params) {
  return '/' + [
    user.body.username, // default to the users org
    project.body.name,
    params.environment,
    project.body.name, // rely on the fact that project name == service name
    user.body.username,
    params.instance // default to instance id 1
  ].join('/');
}

credentials.create = function (session, params, value) {
  return new Promise(function (resolve, reject) {
    if (!(session instanceof Session)) {
      throw new Error('Session must be provided');
    }
    if (!(value instanceof cValue.CredentialValue)) {
      throw new Error('value must be provided');
    }
    if (!params.name ||
        (!params.path &&
         (!params.service || !params.environment ||
          !params.instance || !params.org))) {
      throw new Error('Invalid parameters provided');
    }

    if (params.path && !cpath.validateExp(params.path)) {
      throw new Error('Invalid path provided');
    }

    client.auth(session.token);

    var cpathObj;
    if (params.path) {
      cpathObj = cpath.parseExp(params.path);
    }

    // XXX: Right now the project and service name are the same, so we use
    // that for selecting the service
    var projectName = (cpathObj) ? cpathObj.project : params.service;
    var orgName = (cpathObj) ? cpathObj.org : params.org;
    return Promise.all([
      client.get({ url: '/users/self' }),
      client.get({ url: '/orgs?name=' + orgName })
    ]).then(function (results) {
      var user = results[0] && results[0].body && results[0].body[0];
      var org = results[1] && results[1].body && results[1].body[0];

      if (!user) {
        return reject(new Error('Could not find the user'));
      }

      if (!org) {
        return reject(new Error('Unknown org: ' + orgName));
      }

      var getProject = {
        url: '/projects',
        qs: {
          name: projectName,
          org_id: org.id
        }
      };

      return client.get(getProject).then(function (result) {
        var project = result.body && result.body[0];

        if (!project) {
          return reject(new Error('Unknown project: ' + projectName));
        }

        var pathexp = (cpathObj) ?
          cpathObj.toString() : getPath(user, project, params);

        var getCredential = {
          url: '/credentials',
          qs: {
            name: params.name,
            pathexp: pathexp
          }
        };

        return client.get(getCredential).then(function (credResult) {
          var cred = credResult.body && credResult.body[0];
          var previous = (cred) ? cred.id : null;
          var version = (cred) ? cred.body.version + 1 : 1;

          // Prevent the credential from being unset if it's already unset.
          var curCredValue;
          if (cred && value.body.type === 'undefined') {
            curCredValue = cValue.parse(cred.body.value);

            if (curCredValue.body.type === 'undefined') {
              return reject(new Error('You cannot unset a credential twice'));
            }
          }

          var object = {
            id: utils.id('credential'),
            body: {
              name: params.name,
              project_id: project.id,
              org_id: org.id,
              pathexp: pathexp,
              version: version,
              previous: previous,
              value: value.toString()
            }
          };

          return client.post({
            url: '/credentials',
            json: object
          }).then(function (cResult) {
            resolve(cResult.body);
          });
        });
      });
    }).catch(reject);
  });
};

credentials.get = function (session, params) {
  return new Promise(function (resolve, reject) {
    if (!(session instanceof Session)) {
      throw new Error('Session must be provided');
    }
    if (!params.service || !params.environment ||
        !params.instance || !params.org) {
      throw new Error(
        'Org, Service, environment, and instance must be provided');
    }

    client.auth(session.token);

    return Promise.all([
      client.get({ url: '/users/self' }),
      client.get({ url: '/orgs', qs: { name: params.org } })
    ]).then(function (results) {
      var user = results[0] && results[0].body && results[0].body[0];
      var org = results[1] && results[1].body && results[1].body[0];

      if (!user) {
        return reject(new Error('Could not find user'));
      }

      if (!org) {
        return reject(new Error('Could not find the org: ' + params.org));
      }

      var path = '/' + [
        org.body.name, // default to the users org
        params.service, // service and project have the same name
        params.environment,
        params.service,
        user.body.username,
        params.instance
      ].join('/');

      var getCreds = {
        url: '/credentials',
        qs: {
          path: path
        }
      };
      return client.get(getCreds).then(function (payload) {
        return payload.body;
      });
    })
    .then(resolve)
    .catch(reject);
  });
};
