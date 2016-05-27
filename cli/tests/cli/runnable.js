/* eslint-env mocha */

'use strict';

var sinon = require('sinon');
var assert = require('assert');

var Context = require('../../lib/cli/context');
var Middleware = require('../../lib/cli/middleware');
var Runnable = require('../../lib/cli/runnable');

describe('Runnable', function () {
  var r;
  beforeEach(function () {
    r = new Runnable();
  });

  describe('#hook', function () {
    it('adds pre middleware to list', function () {
      function fn() {}
      r.hook('pre', fn);

      assert.ok(r.preHooks[0] instanceof Middleware);
      assert.strictEqual(r.preHooks.length, 1);
    });

    it('adds post middleware to list', function () {
      function fn() {}
      r.hook('post', fn);

      assert.ok(r.postHooks[0] instanceof Middleware);
      assert.strictEqual(r.postHooks.length, 1);
    });

    it('errors if a fn is not provided', function () {
      assert.throws(function () {
        r.hook('pre', true);
      }, /Middleware must be a function/);
    });

    it('throws err if not pre or post', function () {
      assert.throws(function () {
        r.hook('boo', function fn() {});
      }, /Unknown hook type/);
    });
  });

  describe('#run', function () {
    it('runs all pre middleware and passes in ctx', function () {
      var spy = sinon.spy();
      var spyTwo = sinon.spy();
      var c = new Context({});

      r.hook('pre', spy);
      r.hook('pre', spyTwo);

      return r.runHooks('pre', c).then(function () {
        sinon.assert.calledOnce(spy);
        sinon.assert.calledOnce(spyTwo);

        sinon.assert.calledWith(spy, c);
        sinon.assert.calledWith(spyTwo, c);
      });
    });

    it('runs all post middleware and passes in ctx', function () {
      var spy = sinon.spy();
      var spyTwo = sinon.spy();
      var c = new Context({});

      r.hook('post', spy);
      r.hook('post', spyTwo);

      return r.runHooks('post', c).then(function () {
        sinon.assert.calledOnce(spy);
        sinon.assert.calledOnce(spyTwo);

        sinon.assert.calledWith(spy, c);
        sinon.assert.calledWith(spyTwo, c);
      });
    });


    it('catches and returns middleware error', function () {
      function fn() {
        throw new Error('hi');
      }

      var c = new Context({});

      r.hook('pre', fn);

      return r.runHooks('pre', c).then(function () {
        assert.ok(false, 'Promise should of been rejected');
      }, function (err) {
        assert.ok(err instanceof Error);
        assert.strictEqual(err.message, 'hi');
      });
    });
  });
});
