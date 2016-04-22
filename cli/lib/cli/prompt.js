'use strict';

var _ = require('lodash');
var inquirer = require('inquirer');
var Promise = require('es6-promise').Promise;

function Prompt(stages, opts) {
  if (!_.isFunction(stages)) {
    throw new Error('stages must be a function');
  }
  opts = _.isPlainObject(opts)? opts : {};
  this.stages = stages(this);
  if (!_.isArray(this.stages)) {
    throw new Error('stages must return array');
  }
  this.aggregate = {};
  this._stageAttempts = 0;
  this._stageFailed = false;
  this._maxStageAttempts = opts.maxStageAttempts || 2;
  this._inquirer = inquirer;
}

/**
 * Initiate the prompt, asking all stages
 */
Prompt.prototype.start = function() {
  var self = this;
  var promise = Promise.resolve();
  return promise.then(function() {
    return self.ask(promise, _.keys(self.stages));
  });
};

/**
 * Ask subset of questions in succession
 *
 * @param {object} promise - Promise chain
 * @param {array} stage - Array of stage indexes
 */
Prompt.prototype.ask = function(promise, stage) {
  var self = this;
  return promise.then(function() {
    var questions = self._questions(stage);
    return inquirer.prompt(questions).then(function(result) {
      // Keep track of aggregate answers
      self._aggregate(result);

      // Stage validation failed repeatedly
      if (self._hasFailed()) {
        var retryStage = self._stageFailed;
        self._reset();
        // Recurse until answered properly
        return self.ask(promise, retryStage);
      }

      return self.aggregate;
    });
  });
};

Prompt.prototype._hasFailed = function() {
  return this._stageFailed !== false;
};

Prompt.prototype._questions = function(stage) {
  var self = this;
  var questions = [];
  stage = _.isArray(stage)? stage : [stage];
  _.each(stage, function(stg) {
    questions = questions.concat(self.stages[stg]);
  });
  return questions;
};

Prompt.prototype.failed = function(stageKey, message) {
  this._stageAttempts++;
  if (this._stageAttempts > this._maxStageAttempts) {
    this._stageFailed = stageKey.toString();
    return true;
  }
  return message;
};

Prompt.prototype._aggregate = function(result) {
  this.aggregate = _.extend({}, this.aggregate, result);
};

Prompt.prototype._reset = function() {
  this._stageAttempts = 0;
  this._stageFailed = false;
};

module.exports = Prompt;
