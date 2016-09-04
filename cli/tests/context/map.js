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
    target = new Target({
      path: '/a/b/c/d',
      context: {
        org: 'a',
        project: 'b'
      }
    });
    sandbox = sinon.sandbox.create();

    sandbox.stub(lock, 'acquire').returns(Promise.resolve());
    sandbox.stub(lock, 'release').returns(Promise.resolve());
    sandbox.stub(map, '_writeFile').returns(Promise.resolve());
    sandbox.stub(map, '_rmFile').returns(Promise.resolve());
  });
  afterEach(function () {
    sandbox.restore();
  });

  describe('#unlink', function () {
    it('throws error if config not provided', function () {
      assert.throws(function () {
        map.unlink({});
      }, /Must provide a Target object/);
    });

    it('throws error if target not provided', function () {
      assert.throws(function () {
        map.unlink(cfg, {});
      }, /Must provide a Target object/);
    });

    it('errors if path not in map', function () {
      sandbox.stub(target, 'exists').returns(false);

      return map.unlink(target).then(function () {
        assert.ok(false, 'should error');
      }, function (err) {
        assert.ok(err);
        assert.ok(/Target path is not linked/.test(err.message));
      });
    });

    it('unlinks at exact path', function () {
      return map.unlink(target).then(function () {
        sinon.assert.calledOnce(map._rmFile);
      });
    });
  });
});
