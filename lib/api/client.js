'use strict';

const urlFormat = require('url').format;
const request = require('request');

const client = exports;

const version = require('../util/version');

const METHODS = ['get', 'head', 'post', 'patch', 'del', 'put'];

class Client {

  constructor (registryHostname, sessionToken) {

    registryHostname = registryHostname || 'arigato.tools';

    var registryUrl = client.deriveUrl(registryHostname);
    this.registryUrl = registryUrl;

    // TODO: Come back and have settings to force valid certs etc for
    // development.
    var defaults = {
      baseUrl: registryUrl,
      strictSSL: false, // we need a real development cert
      time: true,
      gzip: true,
      headers: {
        'User-Agent': 'Arigato-CLI '+version.get()
      }
    };

    this.authed = false;
    this._defaultOpts = defaults;

    if (sessionToken) {
      this.authed = true;
      defaults.headers['Authorization'] =
        client.deriveAuthHeader(sessionToken);
    }

    this.r = request.defaults(defaults);
    METHODS.forEach(this._bindVerb.bind(this));
  }

  auth (sessionToken) {
    this.authed = true;
    this.r = this.r.defaults({
      headers: {
        Authorization: client.deriveAuthHeader(sessionToken)
      }
    });
  }

  deauth () {
    this.authed = false;
    this.r = request.defaults(this._defaultOpts);
  }

  _bindVerb (verb) {
    this[verb] = function (opts) {
      return new Promise((resolve, reject) => {
        // It's important that this.r is evaluated at runtime and isn't a
        // reference that is held.. otherwise the funky usage of r.defaults
        // won't work for auth/deauth :)
        this.r[verb](opts, (err, res, body) => {
          if (err) {
            return reject(err);
          }

          try {
            body = (typeof body === 'string') ? JSON.parse(body) : body;
          } catch (parseErr) {
            return reject(new APIError(
              'Invalid JSON returned from registry',
              'invalid_json_error'));
          }

          var success = [200, 201].some((code) => {
            return res.statusCode === code;
          });

          if (!success) {
            return reject(client.extractError(res, body));
          }

          if (Buffer.isBuffer(body)) {
            body = body.toString('utf-8');
          }

          resolve({ res: res, body: body });
        });
      });
    }.bind(this);
  }
}

client.Client = Client;

client.deriveUrl = function (hostname, port) {

  port = (port) ? port : undefined;

  return urlFormat({
    protocol: 'https:',
    slashes: true,
    hostname: hostname,
    port: port
  });
};

client.deriveAuthHeader = function (sessionToken) {
  if (typeof sessionToken !== 'string') {
    throw new TypeError('Session token must be a string');
  }

  return 'Bearer '+sessionToken;
};

class APIError extends Error {
  constructor(messages, code) {
    var message = (Array.isArray(messages)) ? messages[0] : messages;

    message = message || 'No error message provided';
    code = code || 'unknown_api_error';

    super(message, code);

    this.messages = messages;
  }
}

client.APIError = APIError;

client.extractError = function (res, body) {

  var messages;
  var code;

  if (body && typeof body === 'object' && typeof body.error === 'object') {
    code = body.error.type || undefined;
    messages = body.error.message || undefined;
  }

  // TODO: Come back and add some statusCode inspection to guess the error
  // messages and codes..
  return Promise.reject(new APIError(messages, code));
};
