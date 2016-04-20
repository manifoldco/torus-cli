'use strict';

var util = require('util');

var Promise = require('es6-promise').Promise;
var EventRegulator = require('event-regulator').EventRegulator;
var uuid = require('node-uuid');

var Config = require('../config');
var Client = require('./client');

var object = {};
module.exports = object;

function Daemon(cfg) {
  if (!(cfg instanceof Config)) {
    throw new TypeError('Must provide a Config object');
  }

  this.cfg = cfg;
  this.client = new Client(cfg.socketPath);
  this.subscriptions = new EventRegulator([
    [this.client, 'message', this._onMessage.bind(this)],
    [this.client, 'error', this._onError.bind(this), true],
    [this.client, 'close', this._onClose.bind(this), true]
  ]);

  // dictionary for storing current requests.
  this.requests = {};
}
object.Daemon = Daemon;

Daemon.prototype.connect = function () {
  var self = this;
  return new Promise(function (resolve, reject) {
    self.client.connect().then(function() {
      resolve();
    }).catch(reject);
  });
};

Daemon.prototype.disconnect = function () {
  var self = this;
  return new Promise(function (resolve, reject) {
    if (!self.connected()) {
      return reject(new Error('Already disconnected'));
    }

    function onTerminal(err) {
      subscriptions.destroy();
      if (err instanceof Error) {
        return reject(err);
      }

      resolve();
    }

    var subscriptions = new EventRegulator([
      [self.client, 'error', onTerminal],
      [self.client, 'close', onTerminal]
    ]);

    self.client.end();
  });
};

Daemon.prototype.status = function () {
  return this._command('status');
};

Daemon.prototype.get = function () {
  return this._command('get');
};

Daemon.prototype.set = function (data) {
  if (!data && !data.token && !data.passphrase) {
    return Promise.reject(new TypeError(
      'Must provide object with token or passphrase properties'));
  }


  var id = uuid.v4();
  var msg = {
    type: 'request',
    id: id,
    command: 'set',
    body: {
      token: data.token,
      passphrase: data.passphrase
    }
  };
  var p = this._wrap(id);
  this.client.send(msg).catch(p.reject);

  return p;
};

Daemon.prototype.logout = function () {
  return this._command('logout');
};

Daemon.prototype.version = function () {
  return this._command('version');
};

Daemon.prototype._command = function (cmd) {
  var id = uuid.v4();
  var msg = {
    type: 'request',
    id: id,
    command: cmd
  };

  var p = this._wrap(id);
  this.client.send(msg).catch(p.reject);

  return p;
};

Daemon.prototype.connected = function () {
  return this.client.connected();
};

Daemon.prototype._wrap = function (id) {
  var self = this;
  return new Promise(function(resolve, reject) {

    // TODO: Add a timeout here so if we don't get a res back in X seconds we
    // return an error.
    self.requests[id] = {
      resolve: function() {
        delete self.requests[id];
        resolve.apply(null, arguments);
      },
      reject: function() {
        delete self.requests[id];
        reject.apply(null, arguments);
      }
    };
  });
};

Daemon.prototype._onMessage = function (msg) {

  // If the corresponding request doesn't exist then there is a bug in either
  // the daemon or the cli code, so we throw an error.
  var req = this.requests[msg.headers['reply_id']];
  if (!req) {
    throw new Error('Unknown message: '+msg.headers['reply_id']);
  }

  // XXX: Only support replies from the daemon now; anything else gets junked.
  switch (msg.type) {
    case 'reply':
      return req.resolve(msg.body);
    case 'error':
      return req.reject(new DaemonError(msg.body.message));
    default:
      throw new Error('Unsupported message type: '+msg.type);
  }
};

Daemon.prototype._onError = function (err) {
  var self = this;
  Object.keys(this.requests).forEach(function(k) {
     self.requests[k].reject(err);
  });
};

Daemon.prototype._onClose = function () {
  var err = new Error('Socket closed');
  var self = this;
  Object.keys(this.requests).forEach(function(k) {
     self.requests[k].reject(err);
  });
};

function DaemonError (message, code) {
  Error.captureStackTrace(this, this.constructor);
  this.name = this.constructor.name;
  this.message = message || 'Empty Error Message';
  this.code = code || 'EDAEMONERR';
}
util.inherits(DaemonError, Error);
object.DaemonError = DaemonError;
