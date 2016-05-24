'use strict';

var _ = require('lodash');
var inquirer = require('inquirer');
var Promise = require('es6-promise').Promise;

function Prompt(opts) {
  opts = _.isPlainObject(opts)? opts : {};

  if (!_.isFunction(opts.stages)) {
    throw new Error('stages must be a function');
  }

  this.defaults = opts.defaults || {};
  this.stages = opts.stages.apply(this, opts.questionArgs || []);

  if (!_.isArray(this.stages)) {
    throw new Error('stages must return array');
  }

  this.stages = this._defaults(this.stages);

  this.aggregate = {};
  this._stageAttempts = 0;
  this._stageFailed = false;
  this._maxStageAttempts = opts.maxStageAttempts || 1;
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
        self._retryMessage(retryStage);
        // Recurse until answered properly
        return self.ask(promise, retryStage);
      }

      return self.aggregate;
    });
  });
};

Prompt.prototype._retryMessage = function(stageNumber) {
  var stage = this._questions(stageNumber);
  var firstQuestion = stage[0] || {};
  var message = firstQuestion.retryMessage || 'Please try again';
  console.log('');
  console.log(message);
  console.log('');
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

/**
 * Apply default values to stages
 */
Prompt.prototype._defaults = function(stages) {
  var self = this;
  return _.map(stages, function(stage) {
    if (_.isArray(stage)) {
      return self._defaults(stage);
    }
    if (self.defaults[stage.name]) {
      stage.default = self.defaults[stage.name];
    }
    return stage;
  });
};

module.exports = Prompt;
