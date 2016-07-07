'use strict';

var Promise = require('es6-promise').Promise;

var Store = require('./store');
var validate = require('../validate');
var Prompt = require('../cli/prompt');
var client = require('../api/client').create();

var targetMap = require('./map');
var Target = require('./target');

var link = exports;
link.output = {};

var CREATE_ORG_VALUE = 'Create New Org';
var CREATE_PROJECT_VALUE = 'Create New Project';

link.output.success = function (ctx, target) {
  var programName = ctx.program.name;

  console.log('\nYour current working directory and all sub directories have ' +
              'been linked to:\n');

  console.log('Org: ' + target.org);
  console.log('Project: ' + target.project + '\n');

  console.log('Use \'' + programName + ' status\' to display ' +
              'your current working context.\n');
};

link.output.failure = function () {
  console.log('Failed to link your current working directory.');
};

link.execute = function (ctx) {
  return new Promise(function (resolve, reject) {
    client.auth(ctx.session.token);

    var force = ctx.option('force').value;
    var shouldOverwrite = force !== false;

    // Display linked org/project unless --force supplied
    if (!shouldOverwrite && ctx.target.exists()) {
      return resolve(ctx.target);
    }

    // Prompt for org and project
    var store = new Store(client);
    return link._prompt(store).then(function (answers) {
      // Retrieve data objects for supplied values
      return link._retrieveObjects(store, answers);
    }).then(function (objects) {
      // Create target from inputs
      var target = new Target({
        path: ctx.target.path(),
        context: {
          org: objects.org.body.name,
          project: objects.project.body.name
        }
      });

      // Link the current directory
      return targetMap.link(target).then(function () {
        return resolve(target);
      });
    })
    .catch(reject);
  });
};

/**
 * Retrieve or create the org, project objects using the store
 * and answers we got from the user.
 *
 * @param {Store} store
 * @param {Object} answers
 * @returns {Promise}
 */
link._retrieveObjects = function (store, answers) {
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
      return {
        org: org,
        project: project
      };
    });
  });
};

link._prompt = function (store) {
  var prompt = new Prompt({
    stages: link._questions,
    questionArgs: [
      store
    ]
  });

  return prompt.start();
};

link._questions = function (store) {
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
    }
  ];
};
