'use strict';

var fs = require('fs');
var child_process = require('child_process');
var assert = require('assert');

var Promise = require('es6-promise').Promise;
var sinon = require('sinon');

var Config = require('../../lib/config');
var daemon = require('../../lib/daemon');
var Daemon = require('../../lib/daemon/object').Daemon;

describe('Daemon API', function() {

  this.timeout(10000);

  var cfg;
  var sandbox;
  beforeEach(function() {
    cfg = new Config(__dirname);
    sandbox = sinon.sandbox.create();
  });
  afterEach(function() {
    sandbox.restore();
  });

  describe('#get', function() {
    it('returns null if not availabe', function() {
      sandbox.stub(daemon, 'status').returns(Promise.resolve({
        exists: false,
        pid: null
      }));

      return daemon.get(cfg).then(function(d) {
        assert.strictEqual(d, null);
      });
    });

    it('connects if available', function() {
      sandbox.stub(Daemon.prototype, 'connect').returns(Promise.resolve());
      sandbox.stub(daemon, 'status').returns(Promise.resolve({
        exists: true,
        pid: 231
      }));

      return daemon.get(cfg).then(function(d) {
        assert.ok(d instanceof Daemon);

        sinon.assert.calledOnce(d.connect);
      });
    });
  });

  describe('#start', function() {
    it('errors if daemon is already running', function() {
      sandbox.stub(child_process, 'spawn');
      sandbox.stub(daemon, 'status').returns(Promise.resolve({
        exists: true,
        pid: 231
      }));

      return daemon.start(cfg).then(function() {
        assert.ok(false, 'should error');
      }).catch(function(err) {
        sinon.assert.notCalled(child_process.spawn);

        assert.ok(err instanceof Error);
        assert.ok(err.message, 'Daemon is already running');
      });
    });

    it('spawns a child process and returns a daemon object', function() {
      var unrefSpy = sinon.spy();
      sandbox.stub(daemon, 'status').returns(Promise.resolve({
        exists: false,
        pid: null
      }));
      sandbox.stub(child_process, 'spawn', function() {
        return {
          unref: unrefSpy
        };
      });
      sandbox.stub(daemon, 'get').returns(Promise.resolve(new Daemon(cfg)));

      return daemon.start(cfg).then(function(d) {
        sinon.assert.calledOnce(child_process.spawn);
        sinon.assert.calledOnce(unrefSpy);

        assert.ok(d instanceof Daemon);
      });
    });

    it('spawns a child and errors if it cant find daemon', function() {
      var unrefSpy = sinon.spy();
      sandbox.stub(daemon, 'status').returns(Promise.resolve({
        exists: false,
        pid: null
      }));
      sandbox.stub(child_process, 'spawn', function() {
        return {
          unref: unrefSpy
        };
      });
      sandbox.stub(daemon, 'get').returns(Promise.resolve(null));

      return daemon.start(cfg).then(function() {
        assert.ok(false, 'shouldnt happen');
      }).catch(function(err) {
        sinon.assert.calledOnce(child_process.spawn);
        sinon.assert.calledOnce(unrefSpy);

        assert.ok(err instanceof Error);
        assert.strictEqual(err.message, 'Daemon did not start');
      });
    });
  });

  describe('#stop', function() {
    it('errors if daemon is not running', function() {
      sandbox.stub(daemon, 'status').returns(Promise.resolve({
        exists: false,
        pid: null
      }));

      return daemon.stop(cfg).then(function() {
        assert.ok(false, 'should error');
      }).catch(function(err) {
        assert.ok(err instanceof Error);
        assert.strictEqual(err.message, 'Daemon is not running');
      });
    });

    it('sends sigterm if daemon is running', function() {
      sandbox.stub(daemon, 'status').returns(Promise.resolve({
        exists: true,
        pid: 231
      }));
      sandbox.stub(process, 'kill');

      return daemon.stop(cfg).then(function() {
        sinon.assert.calledOnce(process.kill);
        sinon.assert.calledWith(process.kill, 231, 'SIGTERM');
      });
    });

    it('errors if process does not exist', function() {
      var err = new Error('hi');
      err.code = 'ESRCH';

      sandbox.stub(daemon, 'status').returns(Promise.resolve({
        exists: true,
        pid: 231
      }));
      sandbox.stub(process, 'kill').throws(err);

      return daemon.stop(cfg).then(function() {
        assert.ok(false, 'should error');
      }).catch(function(err) {
        assert.ok(err instanceof Error);
        assert.strictEqual(err.message, 'Unknown pid cannot kill: 231');
      });
    });
  });

  describe('#status', function() {

    it('returns null if file does not exist', function() {
      var err = new Error('hi');
      err.code = 'ENOENT';
      sandbox.stub(fs, 'readFile').yields(err);

      return daemon.status(cfg).then(function(status) {
        assert.deepEqual(status, {
          exists: false,
          pid: null
        });
      });
    });

    it('returns error if pid is not valid', function() {
      sandbox.stub(fs, 'readFile').yields(null, 'fsdf');
      return daemon.status(cfg).then(function() {
        assert.ok(false, 'should error');
      }).catch(function(err) {
        assert.ok(err instanceof Error);
        assert.ok(/Invalid pid in file/.test(err.message), 'error msg match');
      });
    });

    it('returns error if kill fails for reason other than ESRCH', function() {
      sandbox.stub(fs, 'readFile').yields(null, '23119');
      sandbox.stub(process, 'kill').throws(new Error('boo'));

      return daemon.status(cfg).then(function() {
        assert.ok(false, 'should error');
      }).catch(function(err) {
        assert.ok(err instanceof Error);
        assert.strictEqual(err.message, 'boo');
      });
    });

    it('handles process no longer existing', function() {
      var err = new Error('hi');
      err.code = 'ESRCH';

      sandbox.stub(fs, 'readFile').yields(null, '23119');
      sandbox.stub(process, 'kill').throws(err);

      return daemon.status(cfg).then(function(status) {
        assert.deepEqual(status, {
          exists: false,
          pid: null
        });
      });
    });

    it('returns pid and status', function() {
      sandbox.stub(fs, 'readFile').yields(null, '23119');
      sandbox.stub(process, 'kill');

      return daemon.status(cfg).then(function(status) {
        sinon.assert.calledWith(process.kill, 23119, 0);
        assert.deepEqual(status, {
          exists: true,
          pid: 23119
        });
      });
    });
  });
});
