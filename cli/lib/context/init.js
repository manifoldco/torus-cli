'use strict';

var Promise = require('es6-promise').Promise;

var Store = require('./store');
var validate = require('../validate');
var Prompt = require('../cli/prompt');
var client = require('../api/client').create();

var targetMap = require('./map');
var Target = require('./target');

var init = exports;
init.output = {};

var CREATE_ORG_VALUE = 'Create New Org';
var CREATE_PROJECT_VALUE = 'Create New Project';
var CREATE_SERVICE_VALUE = 'Create New Service';

init.output.success = function (ctx, objects) {
  var programName = ctx.program.name;
  var target = objects.target;

  console.log('\nYour current working directory and all sub directories have ' +
              'been linked to:\n');
  console.log('Org: ' + target.org);
  console.log('Project: ' + target.project);
  console.log('Environment: ' + target.environment);
  console.log('Service: ' + target.service + '\n');

  console.log('Use \'' + programName + ' status\' to display ' +
              'your current working context.\n');
};

init.output.failure = function () {
  console.log('Failed to link your current working directory.');
};

init.execute = function (ctx) {
  return new Promise(function (resolve, reject) {
    client.auth(ctx.session.token);

    var store = new Store(client);
    init._prompt(store).then(function (answers) {
      return init._retrieveObjects(store, answers);
    }).then(function (objects) {
      return client.get({ url: '/users/self' }).then(function (result) {
        var user = result.body && result.body[0];

        if (!user) {
          throw new Error('Invalid result returned from API');
        }

        objects.user = user;

        var target = new Target(process.cwd(), {
          org: objects.org.body.name,
          project: objects.project.body.name,
          service: objects.service.body.name,

          // XXX: Should we make sure this actually exists??
          environment: 'dev-' + user.body.username
        });
        return targetMap.link(ctx.config, target).then(function () {
          resolve({
            target: target,
            user: user
          });
        });
      });
    })
    .catch(reject);
  });
};

/**
 * Retrieve or create the org, service, and environment objects using the store
 * and answers we got from the user.
 *
 * @param {Store} store
 * @param {Object} anwers
 * @returns {Promise}
 */
init._retrieveObjects = function (store, answers) {
  var getOrg = (answers.org) ?
    Promise.resolve(answers.org) : store.create('org', {
      body: {
        name: answers.orgName
      }
    });

  return getOrg.then(function (org) {
    var getProject = (answers.project) ?
      Promise.resolve(answers.project) : store.create('project', {
        body: {
          name: answers.projectName,
          org_id: org.id
        }
      });

    return getProject.then(function (project) {
      var getService = (answers.service) ?
        Promise.resolve(answers.service) : store.create('service', {
          body: {
            name: answers.serviceName,
            org_id: org.id,
            project_id: project.id
          }
        });

      return getService.then(function (service) {
        return {
          org: org,
          project: project,
          service: service
        };
      });
    });
  });
};

init._prompt = function (store) {
  var prompt = new Prompt({
    stages: init._questions,
    questionArgs: [
      store
    ]
  });

  return prompt.start();
};

init._questions = function (store) {
  // We use state to keep track of the accrued objects so we don't have to keep
  // calling into the store since we can't access the answers hash in a filter.
  var state = {};

  // Use when, filter, and choices along with the store (cache) to figure out
  // which objects we're trying to link.
  //
  // Keep track of the state in the answers hash. If we take in a name we store
  // it in a different property of the answers hash so we can take advantage of
  // validation.
  return [
    {
      type: 'list',
      name: 'org',
      message: 'Select an org',
      choices: function () {
        return store.get('org').then(function (orgs) {
          var choices = orgs.map(function (org) {
            return org.body.name;
          });

          choices.push(CREATE_ORG_VALUE);

          return choices;
        });
      },
      filter: function (orgName) {
        if (orgName === CREATE_ORG_VALUE) {
          return undefined;
        }

        var filter = { body: { name: orgName } };
        return store.get('org', filter).then(function (orgs) {
          if (orgs.length !== 1) {
            throw new Error('org not found: ' + orgName);
          }

          state.org = orgs[0];
          return state.org;
        });
      }
    },
    {
      type: 'input',
      name: 'orgName',
      message: 'New org name?',
      validate: validate.slug,
      when: function (answers) {
        return (answers.org === undefined);
      }
    },
    {
      type: 'list',
      name: 'project',
      message: 'Select a project',
      choices: function (answers) {
        var filter = { body: { org_id: answers.org.id } };
        return store.get('project', filter).then(function (projects) {
          var choices = projects.map(function (project) {
            return project.body.name;
          });

          choices.push(CREATE_PROJECT_VALUE);
          return choices;
        });
      },
      filter: function (projectName) {
        if (projectName === CREATE_PROJECT_VALUE) {
          return undefined;
        }

        var filter = {
          body: {
            org_id: state.org.id,
            name: projectName
          }
        };
        return store.get('project', filter).then(function (projects) {
          if (projects.length !== 1) {
            throw new Error('project not found: ' + projectName);
          }

          state.project = projects[0];
          return state.project;
        });
      },
      when: function (answers) {
        return (answers.org !== undefined);
      }
    },
    {
      type: 'input',
      name: 'projectName',
      message: 'New project name?',
      validate: validate.slug,
      when: function (answers) {
        return (answers.project === undefined);
      }
    },
    {
      type: 'list',
      name: 'service',
      message: 'Select a service',
      choices: function (answers) {
        var filter = {
          body: {
            org_id: answers.org.id,
            project_id: answers.project.id
          }
        };
        return store.get('service', filter).then(function (services) {
          var choices = services.map(function (service) {
            return service.body.name;
          });

          choices.push(CREATE_SERVICE_VALUE);
          return choices;
        });
      },
      filter: function (serviceName) {
        if (serviceName === CREATE_SERVICE_VALUE) {
          return undefined;
        }

        var filter = {
          body: {
            org_id: state.org.id,
            project_id: state.project.id,
            name: serviceName
          }
        };
        return store.get('service', filter).then(function (services) {
          if (services.length !== 1) {
            throw new Error('service not found: ' + serviceName);
          }

          state.service = services[0];
          return state.service;
        });
      },
      when: function (answers) {
        return (answers.project !== undefined);
      }
    },
    {
      type: 'input',
      name: 'serviceName',
      message: 'New service name?',
      validate: validate.slug,
      when: function (answers) {
        return (answers.service === undefined);
      }
    }
  ];
};
