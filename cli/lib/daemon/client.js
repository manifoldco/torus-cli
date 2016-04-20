'use strict';

var util = require('util');
var net = require('net');
var Promise = require('es6-promise').Promise;
var EventEmitter = require('events').EventEmitter;
var EventRegulator = require('event-regulator').EventRegulator;

function proxy (emitter, event) {
  return function () {
    var args = arguments.
    args.unshift(event);

    emitter.emit.apply(emitter, args);
  };
}

/**
 * Daemon Client
 *
 * Wrapper around unix socket for sending and receiving Message objects with a
 * promise based
 */
function Client (socketPath) {
  EventEmitter.call(this);

  if (!socketPath || typeof socketPath !== 'string') {
    throw new TypeError('socketPath string must be provided');
  }

  this.socketPath = socketPath;
  this.socket = new net.Socket();
  this.buf = '';

  this.subscriptions = new EventRegulator([
    [this.socket, 'connect', proxy.bind(this, 'connect')],
    [this.socket, 'drain', proxy.bind(this, 'drain')],
    [this.socket, 'end', proxy.bind(this, 'end')],

    [this.socket, 'data', this._onData.bind(this)],

    [this.socket, 'timeout', this._onTimeout.bind(this), true],
    [this.socket, 'close', this._onClose.bind(this), true],
    [this.socket, 'error', this._onError.bind(this), true]
  ]);
}
util.inherits(Client, EventEmitter);
module.exports = Client;

Client.prototype.connect = function () {
  var self = this;
  return new Promise(function (resolve, reject) {
    self.socket.connect({ path: self.socketPath }, function (err) {
      if (err) {
        return reject(err);
      }

      resolve();
    });
  });
};

Client.prototype.send = function (msg) {
  var self = this;
  return new Promise(function (resolve, reject) {

    if (!msg || !msg.id || !msg.type || !msg.command) {
      return reject(new Error('Message missing required properties'));
    }

    self.socket.write(JSON.stringify(msg), function (err) {
      if (err) {
        return reject(err);
      }

      resolve();
    });
  });
};

Client.prototype.end = function () {
  this.socket.end();
};

Client.prototype.connected = function () {
  return this.socket.writable;
};

Client.prototype._onData = function (buf) {
  this.buf += buf.toString('utf-8');
  if (this.buf.indexOf('\n') === -1) {
    return;
  }

  var parts = this.buf.split('\n');
  this.buf = '';

  var part, msg;
  for (var i = 0; i < parts.length; ++i) {
    part = parts[i];

    // If this it the last piece and it's not an empty string then
    // we know there is more to read before this message is fully
    // received.
    if (i === parts.length - 1 && part.length > 0) {
      this.buf = part;
      break;
    }

    if (part.length === 0) {
      continue;
    }

    try {
      msg = JSON.parse(part);
      this.emit('message', msg);
    } catch (err) {
      return this._onError(new Error(
        'Could not parse message: '+err.message
      ));
    }
  }
};

Client.prototype._onTimeout = function () {
  this.socket.destroy();
  this.emit('error', new Error('Socket timeout'));
};

Client.prototype._onClose = function () {
  this.emit('close');
};

Client.prototype._onError = function (err) {
  this.socket.destroy();
  this.emit('error', err);
};
