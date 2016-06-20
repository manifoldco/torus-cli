/* eslint-env mocha */

'use strict';

var assert = require('assert');

var Promise = require('es6-promise').Promise;
var sinon = require('sinon');

var lock = require('../../lib/util/lock');
var Config = require('../../lib/config');
var Target = require('../../lib/context/target');
var map = require('../../lib/context/map');

describe('targetMap', function () {
  var cfg;
  var target;
  var sandbox;

  beforeEach(function () {
    cfg = new Config(process.cwd(), '0.0.1');
    target = new Target('/a/b/c/d', {
      org: 'a',
      project: 'b',
      service: 'd',
      environment: 'e'
    });
    sandbox = sinon.sandbox.create();

    sandbox.stub(lock, 'acquire').returns(Promise.resolve());
    sandbox.stub(lock, 'release').returns(Promise.resolve());
    sandbox.stub(map, '_writeFile').returns(Promise.resolve());
  });
  afterEach(function () {
    sandbox.restore();
  });

  describe('#link', function () {
    it('throws error if config not provided', function () {
      assert.throws(function () {
        map.link({}, {});
      }, /Must provide a Config object/);
    });

    it('throws error if target not provided', function () {
      assert.throws(function () {
        map.link(cfg, {});
      }, /Must provide a Target object/);
    });

    it('throws if cannot lock file', function () {
      lock.acquire.onCall(0).returns(Promise.reject(new Error('hi')));

      return map.link(cfg, target).then(function () {
        assert.ok(false, 'should error');
      }, function (err) {
        assert.ok(err);
        assert.strictEqual(err.message, 'hi');
      });
    });

    it('errors if map.get errors', function () {
      sandbox.stub(map, 'get').returns(Promise.reject(new Error('woo')));

      return map.link(cfg, target).then(function () {
        assert.ok(false, 'should error');
      }, function (err) {
        assert.ok(err);
        assert.strictEqual(err.message, 'woo');
      });
    });

    it('errors if writeFile fails', function () {
      map._writeFile.onCall(0).returns(Promise.reject(new Error('woo')));

      return map.link(cfg, target).then(function () {
        assert.ok(false, 'should error');
      }, function (err) {
        assert.ok(err);
        assert.strictEqual(err.message, 'woo');
      });
    });

    it('rejects if org/project specified at higher level', function () {
      sandbox.stub(map, 'get').returns(Promise.resolve({
        '/a/b': {
          org: 'a',
          project: 'd',
          service: 'd',
          environment: 'd'
        }
      }));

      return map.link(cfg, target).then(function () {
        assert.ok(false, 'should error');
      }, function (err) {
        assert.ok(err);
        assert.strictEqual(err.message, 'sub-directories cannot link to ' +
          'different org/projects than their parents');
      });
    });

    it('writes the file and returns list', function () {
      sandbox.stub(map, 'get').returns(Promise.resolve({
        '/a/b': {
          org: 'a',
          project: 'b',
          service: 'c',
          environment: 'd'
        },
        '/a/b/c/e': {
          org: 'a',
          project: 'b',
          service: 'e',
          environment: 'd'
        }
      }));

      return map.link(cfg, target).then(function (targets) {
        assert.strictEqual(targets.length, 2);
        assert.strictEqual(targets[0].path, '/a/b/c/d');
        assert.strictEqual(targets[1].path, '/a/b');

        sinon.assert.calledOnce(map._writeFile);
      });
    });
  });

  describe('#unlink', function () {
    it('throws error if config not provided', function () {
      assert.throws(function () {
        map.unlink({}, {});
      }, /Must provide a Config object/);
    });

    it('throws error if target not provided', function () {
      assert.throws(function () {
        map.unlink(cfg, {});
      }, /Must provide a Target object/);
    });

    it('throws if cannot lock file', function () {
      lock.acquire.onCall(0).returns(Promise.reject(new Error('hi')));

      return map.unlink(cfg, target).then(function () {
        assert.ok(false, 'should error');
      }, function (err) {
        assert.ok(err);
        assert.strictEqual(err.message, 'hi');
      });
    });

    it('errors if path not in map', function () {
      sandbox.stub(map, 'get').returns(Promise.resolve({
        '/d/e/f': {
          org: 'd',
          project: 'e',
          service: 'f',
          environment: 'a'
        }
      }));

      return map.unlink(cfg, target).then(function () {
        assert.ok(false, 'should error');
      }, function (err) {
        assert.ok(err);
        assert.ok(/Target path is not linked/.test(err.message));
      });
    });

    it('unlinks at exact path', function () {
      sandbox.stub(map, 'get').returns(Promise.resolve({
        '/a/b/c/d': {
          org: 'a',
          project: 'b',
          service: 'c',
          environment: 'd'
        }
      }));
      return map.unlink(cfg, target).then(function (results) {
        assert.deepEqual(results, []);
      });
    });
  });
});
