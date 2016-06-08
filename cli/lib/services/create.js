'use strict';

var serviceCreate = exports;

var Promise = require('es6-promise').Promise;

var Session = require('../session');
var output = require('../cli/output');
var validate = require('../validate');
var Prompt = require('../cli/prompt');
var client = require('../api/client').create();

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

/**
 * Service prompt questions
 */
serviceCreate.questions = function () {
  return [
    [
      {
        name: 'name',
        message: 'Service name',
        validate: validate.slug
      }
    ]
  ];
};

var validator = validate.build({
  name: validate.slug
});

/**
 * Create prompt for service create
 *
 * @param {object} ctx - Command context
 */
serviceCreate.execute = function (ctx) {
  return new Promise(function (resolve, reject) {
    var retrieveInput;

    var options = ctx.options;
    var orgName = options && options.org && options.org.value;

    if (!orgName) {
      throw new Error('--org is (temporarily) required.');
    }

    var name = ctx.params[0];
    if (name) {
      var errors = validator({ name: name });
      if (errors.length > 0) {
        retrieveInput = Promise.reject(errors[0]);
      } else {
        retrieveInput = Promise.resolve({ name: name });
      }
    } else {
      retrieveInput = serviceCreate._prompt();
    }

    return retrieveInput.then(function (userInput) {
      return serviceCreate._execute(ctx.session, userInput, orgName);
    }).then(resolve).catch(reject);
  });
};

/**
 * Attempt to create service with supplied input
 *
 * @param {object} session - Session object
 * @param {object} userInput
 */
serviceCreate._execute = function (session, userInput, orgName) {
  return new Promise(function (resolve, reject) {
    if (!(session instanceof Session)) {
      throw new TypeError('Session object missing on Context');
    }

    client.auth(session.token);

    // Projects and services have the same name for now; standardized until
    // projects are part of the UI we reveal to users through the CLI.
    var project;
    var org;
    client.get({
      url: '/orgs',
      qs: { name: orgName }
    })
    .then(function (res) {
      org = res.body && res.body[0];

      if (!org) {
        throw new Error('Project was not created; invalid org provided');
      }

      return client.post({
        url: '/projects',
        json: {
          body: {
            name: userInput.name,
            org_id: org.id
          }
        }
      });
    }).then(function (res) {
      if (!res.body || res.body.length !== 1) {
        throw new Error('Project was not created; invalid body returned');
      }

      project = res.body[0];
      return client.post({
        url: '/services',
        json: {
          body: {
            name: project.body.name,
            project_id: project.id,
            org_id: org.id
          }
        }
      });
    })
    .then(function (res) {
      if (!res.body || res.body.length !== 1) {
        throw new Error('Service was not created; invalid body returned');
      }

      var service = res.body[0];
      resolve({
        project: project,
        service: service
      });
    })
    .catch(reject);
  });
};

/**
 * Create prompt promise
 */
serviceCreate._prompt = function () {
  var prompt = new Prompt({
    stages: serviceCreate.questions
  });

  return prompt.start();
};
