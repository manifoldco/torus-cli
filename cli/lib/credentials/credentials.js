'use strict';

var _ = require('lodash');
var Promise = require('es6-promise').Promise;

var cpath = require('common/cpath');
var cValue = require('./value');

var credentials = exports;

function getPath(params) {
  return '/' + [
    params.org, // default to the users org
    params.project,
    params.environment,
    params.service,
    params.identity,
    params.instance // default to instance id 1
  ].join('/');
}

credentials.create = function (api, params, value) {
  return new Promise(function (resolve, reject) {
    if (!(value instanceof cValue.CredentialValue)) {
      throw new Error('value must be provided');
    }
    if (!params.name ||
        (!params.path &&
         (!params.org || !params.project || !params.service ||
          !params.environment || !params.instance))) {
      throw new Error('Invalid parameters provided');
    }

    if (params.path && !cpath.validateExp(params.path)) {
      throw new Error('Invalid path provided');
    }

    var cpathObj;
    if (params.path) {
      cpathObj = cpath.parseExp(params.path);
    }

    var projectName = (cpathObj) ? cpathObj.project : params.project;
    var orgName = (cpathObj) ? cpathObj.org : params.org;
    return Promise.all([
      api.users.self(),
      api.orgs.get({ name: orgName })
    ]).then(function (results) {
      var user = results[0];
      var org = results[1] && results[1][0];

      if (!user) {
        return reject(new Error('Could not find the user'));
      }

      if (!org) {
        return reject(new Error('Unknown org: ' + orgName));
      }

      var qs = { name: projectName, org_id: org.id };
      return api.projects.get(qs).then(function (result) {
        var project = result[0];
        if (!project) {
          return reject(new Error('Unknown project: ' + projectName));
        }

        var pathexp = (cpathObj) ? cpathObj.toString() : getPath(params);
        var data = {
          name: params.name,
          project_id: project.id,
          org_id: org.id,
          pathexp: pathexp,
          value: value.toString()
        };

        return api.credentials.create(data, function (event) {
          console.log(event.message);
        });
      });
    })
    .then(resolve)
    .catch(reject);
  });
};

credentials.get = function (api, params) {
  return new Promise(function (resolve, reject) {
    if (!params.project || !params.service || !params.environment ||
        !params.instance || !params.org) {
      throw new Error(
        'Org, project, service, environment, and instance must be provided');
    }

    var path;
    return Promise.all([
      api.users.self(),
      api.orgs.get({ name: params.org })
    ]).then(function (results) {
      var user = results[0];
      var org = results[1] && results[1][0];

      if (!user) {
        return reject(new Error('Could not find user'));
      }

      if (!org) {
        return reject(new Error('Could not find the org: ' + params.org));
      }

      path = '/' + [
        org.body.name,
        params.project,
        params.environment,
        params.service,
        user.body.username,
        params.instance
      ].join('/');

      return api.credentials.get({ path: path });
    })
    .then(function (creds) {
      // TODO: Move this logic into the daemon.
      //
      // The daemon will return to us all credentials in all keyrings; some of
      // these may collide in the credential `name` space (since name ==
      // process env variable).
      //
      // Therefore, we need to collapse based on path specificity!
      var nameMap = {};
      var name;
      var cred;
      var cv;
      for (var i = 0; i < creds.length; ++i) {
        cred = creds[i];
        name = cred.body.name;

        // If cred has been unset then ignore it.
        cv = cValue.parse(cred.body.value);
        if (cv.body.type === 'undefined') {
          continue;
        }

        if (!nameMap[name]) {
          nameMap[name] = cred;
          continue;
        }

        // Figure out which is the most specific path
        if (cpath.compare(
          nameMap[name].body.pathexp, cred.body.pathexp) === -1) {
          nameMap[name] = cred;
        }
      }

      resolve({
        credentials: _.values(nameMap),
        path: path
      });
    })
    .catch(reject);
  });
};
