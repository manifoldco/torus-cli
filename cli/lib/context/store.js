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
 * It's been desigend with the `ag init` flow in mind, nothing more.
 *
 * @constructor
 * @param {Client} client an api client
 */
function Store(client) {
  this.state = {};
  this.types = [
    'org',
    'project',
    'service'
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

var CREATE_MAP = {
  org: {
    url: '/orgs'
  },
  project: {
    url: '/projects'
  },
  service: {
    url: '/services'
  }
};

Store.prototype.create = function (type, data) {
  var create = {
    url: CREATE_MAP[type].url,
    json: data
  };

  var self = this;
  return this.client.post(create).then(function (result) {
    var object = result && result.body && result.body[0];
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

// XXX: This only supports querying against properties on the body
var GET_MAP = {
  org: {
    url: '/orgs',
    qs: []
  },
  project: {
    url: '/projects',
    qs: [
      'org_id'
    ]
  },
  service: {
    url: '/services',
    qs: [
      'org_id',
      'project_id'
    ]
  }
};

// XXX: We assume _initialize only gets run once!
Store.prototype._initialize = function (type, filter) {
  if (!GET_MAP[type]) {
    throw new TypeError('Invalid get type for _initialize: ' + type);
  }

  var get = {
    url: GET_MAP[type].url,
    qs: {}
  };

  GET_MAP[type].qs.forEach(function (property) {
    var value = _.get(filter, 'body.' + property);
    if (!value) {
      throw new TypeError('Missing required property on filter: ' + property);
    }

    get.qs[property] = value;
  });

  var self = this;
  return this.client.get(get).then(function (result) {
    var objects = result && result.body;
    if (!objects) {
      throw new Error('Invalid result returned from server');
    }

    self.state[type] = objects;

    return objects;
  });
};
