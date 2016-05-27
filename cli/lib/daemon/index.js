'use strict';

var fs = require('fs');
var path = require('path');
var childProcess = require('child_process');

var _ = require('lodash');
var Promise = require('es6-promise').Promise;
var Backoff = require('backo');

var Daemon = require('./object').Daemon;
var Config = require('../config');

var daemon = exports;

// The daemon binary is placed in the cli's `bin` folder during the build
// process.
daemon.DAEMON_PATH = path.join(__dirname, '../../bin/ag-daemon');

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

    return daemon.status(cfg).then(function (status) {
      if (!status.exists) {
        return resolve(null);
      }


      var d = new Daemon(cfg);
      return d.connect().then(function () {
        resolve(d);
      }).catch(reject);
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

    function retry(backoff) {
      if (backoff.attempts > 3) {
        return reject(new Error('Daemon did not start'));
      }

      return setTimeout(function () {
        daemon.get(cfg).then(function (d) {
          if (!d) {
            return retry(backoff);
          }

          return resolve(d);
        }).catch(reject);
      }, backoff.duration());
    }

    return daemon.status(cfg).then(function (status) {
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
      var child = childProcess.spawn(daemon.DAEMON_PATH, [], opt);
      child.unref();

      // XXX: Wait a second before trying to connect to the daemon
      var backoff = new Backoff({
        min: 100, max: 1000, jitter: 0.75, factor: 2 });
      return retry(backoff);
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

    return daemon.status(cfg).then(function (status) {
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
          return reject(new Error('Unknown pid cannot kill: ' + status.pid));
        }

        return reject(err);
      }

      return resolve();
    });
  });
};

daemon.status = function (cfg) {
  return new Promise(function (resolve, reject) {
    fs.readFile(cfg.pidPath, 'utf-8', function (err, pid) {
      if (err) {
        if (err && err.code === 'ENOENT') {
          return resolve({ exists: false, pid: null });
        }

        return reject(err);
      }

      pid = parseInt(pid, 10);
      if (_.isNaN(pid)) {
        return reject(new Error('Invalid pid in file ' + cfg.pidPath));
      }

      var exists = true;
      try {
        // Process.kill sends a signal, if you set the vlaue to 0 it
        // just checks if the process exists.
        //
        // An error is thrown if the process does not exist.
        process.kill(pid, 0);
      } catch (e) {
        if (e.code !== 'ESRCH') {
          return reject(e);
        }

        exists = false;
        pid = null;
      }

      return resolve({ exists: exists, pid: pid });
    });
  });
};
