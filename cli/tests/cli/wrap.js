'use strict';

var assert = require('assert');
var Promise = require('es6-promise').Promise;

var wrap = require('../../lib/cli/wrap');

describe('wrap', function() {
  
  it('returns promise w/ value if no promise from fn', function() {
    var r = wrap(function() {
      return 42;
    });

    assert.ok(r instanceof Promise);
    return r.then(function(v) {
      assert.strictEqual(v, 42);
    });
  });

  it('wraps errors and rejects with promise', function () {
    var r = wrap(function() {
      throw new Error('hi');
    });

    assert.ok(r instanceof Promise);

    return r.then(function() {
      assert.ok(false, 'Promise should be rejected');
    }).catch(function(err) {
      assert.ok(err instanceof Error);
      assert.strictEqual(err.message, 'hi');
    });
  });

  it('proxies the error back', function () {
    var p = Promise.reject(new Error('hi'));
    var r = wrap(function() {
      return p;
    });

    assert.strictEqual(r, p);
    assert.ok(r instanceof Promise);

    return r.then(function() {
      assert.ok(false, 'Promise should be rejected');
    }).catch(function(err) {
      assert.ok(err instanceof Error);
      assert.strictEqual(err.message, 'hi');
    });
  });

  it('proxies the value back', function() {
    var p = Promise.resolve({ a: 'hi' });
    var r = wrap(function() {
      return p;
    });

    return r.then(function(v) {
      assert.strictEqual(v.a, 'hi');
    });
  });
});
