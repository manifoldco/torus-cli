'use strict';

var client = exports;

var _ = require('lodash');
var request = require('request');
var Promise = require('es6-promise').Promise;

var CLI_VERSION = require('../../package.json').version;

var HTTP_VERBS = [
  'post',
  'get',
  'put',
  'patch',
  'delete'
];

/**
 * Arigato API client
 *
 * @param {object} opts
 */
function Client(opts) {
  opts = opts || {};
  this.proxyEndpoint = opts.socketUrl + '/proxy';
  this.v1Endpoint = opts.socketUrl + '/v1';
  this.version = {
    cli: CLI_VERSION,
    api: opts.apiVersion || null
  };
  this._initialize();
}

Client.prototype.attach = function (name, module) {
  this[name] = {};

  var target = this[name];
  var c = this;
  _.each(module, function (method, apiName) {
    if (typeof method !== 'function') {
      return;
    }

    target[apiName] = method.bind(module, c);
  });
};

/**
 * Initialize verb-specific methods
 */
Client.prototype._initialize = function () {
  var self = this;
  HTTP_VERBS.forEach(function (verb) {
    self[verb] = self._req.bind(self, verb);
  });
};

/**
 * Perform request
 *
 * @param {string} verb
 * @param {object} opts
 */
Client.prototype._req = function (verb, opts, isV1) {
  var self = this;
  return new Promise(function (resolve, reject) {
    if (opts.url.indexOf(':') > -1) {
      if (!opts.params || Object.keys(opts.params).length === 0) {
        throw new Error('Request to ' + opts.url + ' requires params');
      }

      _.each(opts.params, function (value, name) {
        opts.url = opts.url.replace(':' + name, value);
      });
    }

    opts = _.extend({}, opts, {
      method: verb,
      url: (isV1 ? self.v1Endpoint : self.proxyEndpoint) + opts.url,
      headers: self._headers(opts),
      time: true,
      gzip: true,
      timeout: 3000
    });

    request(opts, function (err, res, body) {
      if (err) {
        return reject(err);
      }

      try {
        body = _.isString(body) && body.length > 0 ? JSON.parse(body) : body;
        body = _.isString(body) && body.length === 0 ? null : body;
        res.body = body;
      } catch (e) {
        return reject(new Error('invalid json returned from API'));
      }

      if (!self._statusSuccess(res)) {
        return reject(self._extractError(body));
      }

      return resolve(res);
    });
  });
};

/**
 * Extend headers with defaults
 *
 * @param {object} opts
 */
Client.prototype._headers = function (opts) {
  var headers = {
    'User-Agent': 'Arigato CLI ' + this.version.cli,
    'Content-Type': 'application/json',
    Host: 'arigato.tools'
  };

  if (this.version.api) {
    headers['x-arigato-version'] = this.version.api;
  }
  return _.extend({}, opts.headers, headers);
};

/**
 * Identify error from body
 *
 * @param {object} body
 */
Client.prototype._extractError = function (body) {
  var err = new Error(body.error);
  err.type = body.type;
  return err;
};

/**
 * Return true if 2xx status code
 *
 * @param {object} res
 */
Client.prototype._statusSuccess = function (res) {
  return Math.floor(res.statusCode / 100) === 2;
};

client.Client = Client;
