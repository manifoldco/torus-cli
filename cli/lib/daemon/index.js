'use strict';

var fs = require('fs');
var path = require('path');
var child_process = require('child_process');

var _ = require('lodash');
var Promise = require('es6-promise').Promise;

var Daemon = require('./object').Daemon;
var Config = require('../config');

var daemon = exports;
daemon.DAEMON_PATH = path.join(__dirname, '../../../daemon/ag-daemon');

/**
 * Retrieves a deamon object, returns null if the daemon is not running
 *
 * @param cfg {Config}
 * @returns Promise
 */
daemon.get = function (cfg) {
  return new Promise(function (resolve, reject) {
    if (!(cfg instanceof Config)) {
      return reject(new TypeError('cfg must be a Config object'));
    }

    daemon.status(cfg).then(function(status) {
      if (!status.exists) {
        return resolve(null);
      }


      var d = new Daemon(cfg);
      return d.connect().then(function() {
        resolve(d);
      });
    }).catch(reject);
  });
};

/**
 * Starts the daemon and returns a daemon instance. Rejects the promise with an
 * error if the daemon is already running
 *
 * @param cfg {Config}
 * @returns Promise
 */
daemon.start = function (cfg) {
  return new Promise(function (resolve, reject) {
    if (!(cfg instanceof Config)) {
      return reject(new TypeError('cfg must be a Config object'));
    }

    daemon.status(cfg).then(function(status) {
      if (status.exists) {
        return reject(new Error('Daemon is already running'));
      }

      var env = _.clone(process.env);
      env.ARIGATO_ROOT = cfg.arigatoRoot;

      var opt = {
        stdio: ['ignore', 'ignore', 'ignore'],
        env: env,
        cwd: cfg.arigatoRoot,
        detached: true
      };

      // TODO: We should be logging the pid and other information
      var child = child_process.spawn(daemon.DAEMON_PATH, [], opt);
      child.unref();

      // XXX: Wait a second before trying to connect to the daemon
      setTimeout(function() {
        daemon.get(cfg).then(function(d) {
          if (!d) {
            return reject(new Error('Daemon did not start'));
          }

          resolve(d);
        }).catch(reject);
      }, 1000);
    }).catch(reject);
  });
};

/**
 * Stops the currently running daemon. Rejects the promise with an error if the
 * daemon is not running.
 *
 * @param cfg {Config}
 * @returns Promise
 */
daemon.stop = function (cfg) {
  return new Promise(function (resolve, reject) {
    if (!(cfg instanceof Config)) {
      return reject(new TypeError('cfg must be a Config object'));
    }

    daemon.status(cfg).then(function(status) {
      if (!status.exists) {
        return reject(new Error('Daemon is not running'));
      }

      // Send a SIGTERM and let the daemon shutdown gracefully.
      //
      // TODO: Add a grace period, if it's still alive after N seconds then
      // send a SIGKILL.
      try {
        process.kill(status.pid, 'SIGTERM');
      } catch (err) {
        if (err.code === 'ESRCH') {
          return reject(new Error('Unknown pid cannot kill: '+status.pid));
        }

        return reject(err);
      }

      resolve();
    });
  });
};

daemon.status = function (cfg) {
  return new Promise(function(resolve, reject) {
    fs.readFile(cfg.pidPath, 'utf-8', function (err, pid) {
      if (err) {
        if (err && err.code === 'ENOENT') {
          return resolve({ exists: false, pid: null });
        }

        return reject(err);
      }

      pid = parseInt(pid, 10);
      if (_.isNaN(pid)) {
        return reject(new Error('Invalid pid in file '+cfg.pidPath));
      }

      var exists = true;
      try {
        // Process.kill sends a signal, if you set the vlaue to 0 it
        // just checks if the process exists.
        //
        // An error is thrown if the process does not exist.
        process.kill(pid, 0);
      } catch (err) {
        if (err.code !== 'ESRCH') {
          return reject(err);
        }

        exists = false;
        pid = null;
      }

      resolve({ exists: exists, pid: pid });
    });
  });
};
