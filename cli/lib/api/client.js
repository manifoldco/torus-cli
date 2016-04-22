'use strict';

var _ = require('lodash');
var request = require('request');
var Promise = require('es6-promise').Promise;

var HTTP_VERBS = [
  'post',
  'get',
  'put',
  'patch',
  'delete'
];

var CLI_VERSION = require('../../../package.json').version;

function Client(opts) {
  opts = opts || {};
  this.endpoint = opts.endpoint || 'https://arigato.tools';
  this.authToken = opts.authToken || null;
  this.version = {
    cli: CLI_VERSION,
    api: opts.apiVersion || null
  };
  this._initialize();
}

/**
 * Set authtoken property
 *
 * @param {string} authToken
 */
Client.prototype.auth = function(authToken) {
  if (typeof authToken !== 'string') {
    throw new Error('auth token must be a string');
  } 

  this.authToken = authToken;
};

/**
 * Initialize verb-specific methods
 */
Client.prototype._initialize = function() {
  var self = this;
  HTTP_VERBS.forEach(function(verb) {
    self[verb] = self._req.bind(self, verb);
  });
};

/**
 * Perform request
 *
 * @param {string} verb
 * @param {object} opts
 */
Client.prototype._req = function(verb, opts) {
  var self = this;
  return new Promise(function(resolve, reject) {
    opts = _.extend({}, opts, {
      method: verb,
      url: self.endpoint + opts.url,
      headers: self._headers(opts),
      strictSSL: false, // TODO: proper development cert
      time: true,
      gzip: true,
    });

    request(opts, function(err, res, body) {
      if (err) {
        return reject(err);
      }

      try {
        body = (typeof body === 'string')? JSON.parse(body) : body;
        res.body = body;
      } catch(err) {
        return reject(new Error('invalid json returned from API'));
      }

      if (!self._statusSuccess(res)) {
        return reject(self._extractError(body));
      }

      resolve(res);
    });
  });
};

/**
 * Extend headers with defaults
 *
 * @param {object} opts
 */
Client.prototype._headers = function(opts) {
  var headers = {
    'User-Agent': 'Arigato CLI ' + this.version.cli,
    'Content-Type': 'application/json'
  };

  if (this.version.api) {
    headers['x-arigato-version'] = this.version.api;
  }
  if (this.authToken) {
    headers['Authorization'] = 'Bearer ' + this.authToken;
  }
  return _.extend({}, opts.headers, headers);
};

/**
 * Identify error from body
 *
 * @param {object} body
 */
Client.prototype._extractError = function(body) {
  var err = new Error(body.error);
  err.type = body.type;
  return err;
};

/**
 * Return true if 2xx status code
 *
 * @param {object} res
 */
Client.prototype._statusSuccess = function(res) {
  return Math.floor(res.statusCode / 100) === 2;
};

module.exports = Client;
