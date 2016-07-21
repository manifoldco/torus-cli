'use strict';

var _ = require('lodash');

var EFFECTS = {
  ALLOW: 'allow',
  DENY: 'deny'
};

var ACTIONS = {
  CREATE: 'create',
  READ: 'read',
  UPDATE: 'update',
  DELETE: 'delete',
  LIST: 'list'
};

var ACTIONS_SHORTHAND = _.reduce(ACTIONS, function (a, action) {
  a[_.first(action)] = action;
  return a;
}, {});

function Statement(effect) {
  this.effect = effect;
}

Statement.EFFECTS = EFFECTS;
Statement.ACTIONS = ACTIONS;
Statement.ACTIONS_SHORTHAND = ACTIONS_SHORTHAND;

Statement.prototype.parseActions = function (actionsShorthand) {
  return _.map(actionsShorthand.split(''), function (char) {
    var action = ACTIONS_SHORTHAND[char];

    if (!action) {
      throw new Error('invalid action shorthand provided');
    }

    return action;
  });
};

Statement.prototype.setResource = function (resource) {
  this.resource = resource;
};

Statement.prototype.setActions = function (actions) {
  if (!_.isArray(actions) && !_.isString(actions)) {
    throw new Error('invalid actions provided');
  }

  if (_.isArray(actions)) {
    actions = _.intersection(actions, _.values(ACTIONS));
  }

  if (_.isString(actions)) {
    actions = this.parseActions(actions);
  }

  if (actions.length < 1) {
    throw new Error('no valid actions provided');
  }

  this.actions = actions;
};

function Policy(name, description) {
  this.name = name || '';
  this.description = description || '';
  this.statements = [];
}

Policy.prototype.addStatement = function (statement) {
  this.statements.push(statement);
};

module.exports = {
  Policy: Policy,
  Statement: Statement
};
