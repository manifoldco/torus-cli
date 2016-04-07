'use strict';

var sinon = require('sinon');
var assert = require('assert');

var Context = require('../../lib/cli/context');
var Middleware = require('../../lib/cli/middleware');
var Runnable = require('../../lib/cli/runnable');

describe('Runnable', function() {
  
  var r;
  beforeEach(function() {
    r = new Runnable();
  });

  describe('#use', function () {
    it('adds middleware to list', function() {
      function fn () {}
      r.use(fn);

      assert.ok(r.middleware[0] instanceof Middleware);
      assert.strictEqual(r.middleware.length, 1);
    });

    it('errors if a fn is not provided', function() {
      assert.throws(function() {
        r.use(true);
      }, /Middleware must be a function/);
    });
  });

  describe('#run', function() {
    it('runs all middleware and passes in ctx', function() {
      var spy = sinon.spy();
      var spyTwo = sinon.spy();
      var c = new Context({});

      r.use(spy);
      r.use(spyTwo);

      return r.runMiddleware(c).then(function(results) {
        sinon.assert.calledOnce(spy);
        sinon.assert.calledOnce(spyTwo);

        sinon.assert.calledWith(spy, c);
        sinon.assert.calledWith(spyTwo, c);
      });
    });

    it('catches and returns middleware error', function() {
      function fn () {
        throw new Error('hi');
      }

      var c = new Context({});

      r.use(fn);

      return r.runMiddleware(c).then(function() {
        assert.ok(false, 'Promise should of been rejected');
      }, function(err) {
        assert.ok(err instanceof Error);
        assert.strictEqual(err.message, 'hi');
      });
    });
  });
});
