'use strict';

var assert = require('assert');
var sinon = require('sinon');

var Context = require('../../lib/cli/context');
var Middleware = require('../../lib/cli/middleware');

describe('Middleware', function() {

  var m;
  describe('constructor', function() {
    it('throws error if a fn is not provided', function() {
      assert.throws(function() {
        m = new Middleware('a');
      }, /Middleware must be a function/);
    });

    it('constructs with a function', function() {
      m = new Middleware(function(ctx) {
        console.log('hi');
      });

      assert.ok(m instanceof Middleware);
      assert.ok(m.fn);
    });
  });

  describe('#run', function() {

    var spy;
    beforeEach(function() {
      spy = sinon.spy();
    });

    it('runs the function and passes in a context', function() {
      var c = new Context({});
      var m = new Middleware(spy);

      return m.run(c).then(function() {
        sinon.assert.calledOnce(spy);
        sinon.assert.calledWith(spy, c);
      });
    });
  });
});
