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

init.output.success = function (objects) {
  var target = objects.target;
  var user = objects.user;

  console.log('\n' + process.cwd(), 'has been associated with /' + [
    target.org,
    target.project,
    target.environment,
    target.service,
    user.body.username,
    '1'
  ].join('/') + '\n');

  console.log('All sub-directories will be associated with this org, project,' +
    ' and service');
};

init.output.failure = function () {
  console.log('Failed to associate directory');
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
      message: 'The organization this code base is tied too',
      choices: function () {
        return store.get('org').then(function (orgs) {
          var choices = orgs.map(function (org) {
            return org.body.name;
          });

          choices.push('Create New Organization');

          return choices;
        });
      },
      filter: function (orgName) {
        if (orgName === 'Create New Organization') {
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
      message: 'What\'s the name of the new organization?',
      validate: validate.slug,
      when: function (answers) {
        return (answers.org === undefined);
      }
    },
    {
      type: 'list',
      name: 'project',
      message: 'Which project does this codebase belong too?',
      choices: function (answers) {
        var filter = { body: { org_id: answers.org.id } };
        return store.get('project', filter).then(function (projects) {
          var choices = projects.map(function (project) {
            return project.body.name;
          });

          choices.push('Create New Project');
          return choices;
        });
      },
      filter: function (projectName) {
        if (projectName === 'Create New Project') {
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
      message: 'What\'s the name of the new project?',
      validate: validate.slug,
      when: function (answers) {
        return (answers.project === undefined);
      }
    },
    {
      type: 'list',
      name: 'service',
      message: 'What service does this codebase belong too?',
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

          choices.push('Create New Service');
          return choices;
        });
      },
      filter: function (serviceName) {
        if (serviceName === 'Create New Service') {
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

          return services[0];
        });
      },
      when: function (answers) {
        return (answers.project !== undefined);
      }
    },
    {
      type: 'input',
      name: 'serviceName',
      message: 'What\'s the name of the new service?',
      validate: validate.slug,
      when: function (answers) {
        return (answers.service === undefined);
      }
    }
  ];
};
