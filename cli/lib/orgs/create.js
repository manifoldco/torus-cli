'use strict';

var orgCreate = exports;

var Promise = require('es6-promise').Promise;
var _ = require('lodash');

var output = require('../cli/output');
var validate = require('../validate');
var Prompt = require('../cli/prompt');

orgCreate.output = {};

orgCreate.output.success = output.create(function () {
  console.log('Org created');
});

orgCreate.output.failure = output.create(function () {
  console.log('Org creation failed, please try again');
});

var validator = validate.build({
  name: validate.slug
});

/**
 * Create prompt for org create
 *
 * @param {object} ctx - Command context
 */
orgCreate.execute = function (ctx) {
  return new Promise(function (resolve, reject) {
    var data = {
      name: ctx.params[0]
    };

    if (data.name) {
      var errors = validator(data);

      if (errors.length > 0) {
        return reject(errors[0]);
      }
    }

    var createOrg = orgCreate._prompt(data).then(function (input) {
      return ctx.api.orgs.create(input);
    });

    return createOrg.then(resolve).catch(reject);
  });
};

/**
 * Create prompt promise
 */
orgCreate._prompt = function (defaults) {
  var prompt = new Prompt({
    stages: orgCreate._questions,
    defaults: defaults
  });

  return prompt.start().then(function (answers) {
    return _.omitBy(_.extend({}, defaults, answers), _.isUndefined);
  });
};

orgCreate._questions = function () {
  return [
    [
      {
        name: 'name',
        message: 'Org name',
        validate: validate.slug
      }
    ]
  ];
};
