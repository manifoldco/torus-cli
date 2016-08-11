'use strict';

var _ = require('lodash');
var Promise = require('es6-promise').Promise;

/**
 * Simple wrapper around the Registry API for caching different types of
 * objects and then interacting with those objects (retrieving and creating).
 *
 * Does not implement a reactive event-based interface as we don't need one
 * right now (nothing is event driven or interactive).
 *
 * It also won't update based on changes made to the underlying data on the
 * server by another user.
 *
 * It's been desigend with the `ag link` flow in mind, nothing more.
 *
 * @constructor
 * @param {Client} client an api client
 */
function Store(client) {
  this.state = {};
  this.types = [
    'orgs',
    'projects',
    'services'
  ];

  this.client = client;
}
module.exports = Store;

Store.prototype.get = function (type, filter) {
  if (this.types.indexOf(type) === -1) {
    throw new TypeError('unknown object type: ' + type);
  }

  var getCurrentState = (!this.state[type] || this.state[type].length === 0) ?
    this._initialize(type, filter) : Promise.resolve(this.state[type]);

  return getCurrentState.then(function (data) {
    return _.filter(data, filter);
  });
};

Store.prototype.create = function (type, data) {
  if (this.types.indexOf(type) === -1) {
    throw new TypeError('unknown object type: ' + type);
  }

  var self = this;
  return this.client[type].create(data).then(function (object) {
    if (!object) {
      throw new Error('Invalid result returned from API');
    }

    if (!self.state[type]) {
      self.state[type] = [];
    }

    self.state[type].push(object);
    return object;
  });
};

// XXX: We assume _initialize only gets run once!
Store.prototype._initialize = function (type, filter) {
  var self = this;
  var query = (filter && filter.body) ? filter.body : {};
  return this.client[type].get(query).then(function (objects) {
    if (!objects) {
      throw new Error('Invalid result returned from server');
    }

    self.state[type] = objects;

    return objects;
  });
};
