/* eslint-env mocha */

'use strict';

var assert = require('assert');
var Promise = require('es6-promise').Promise;

var wrap = require('../../lib/cli/wrap');

describe('wrap', function () {
  it('returns promise w/ value if no promise from fn', function () {
    var r = wrap(function () {
      return 42;
    });

    assert.ok(r instanceof Promise);
    return r.then(function (v) {
      assert.strictEqual(v, 42);
    });
  });

  it('wraps errors and rejects with promise', function () {
    var r = wrap(function () {
      throw new Error('hi');
    });

    assert.ok(r instanceof Promise);

    return r.then(function () {
      assert.ok(false, 'Promise should be rejected');
    }).catch(function (err) {
      assert.ok(err instanceof Error);
      assert.strictEqual(err.message, 'hi');
    });
  });

  it('proxies the error back', function () {
    var p = Promise.reject(new Error('hi'));
    var r = wrap(function () {
      return p;
    });

    assert.strictEqual(r, p);
    assert.ok(r instanceof Promise);

    return r.then(function () {
      assert.ok(false, 'Promise should be rejected');
    }).catch(function (err) {
      assert.ok(err instanceof Error);
      assert.strictEqual(err.message, 'hi');
    });
  });

  it('returns 0 if no value provided', function () {
    var r = wrap(function () {});
    return r.then(function (v) {
      assert.strictEqual(v, 0);
    });
  });

  it('returns 1 if false provided', function () {
    var r = wrap(function () {
      return false;
    });

    return r.then(function (v) {
      assert.strictEqual(v, 1);
    });
  });

  it('returns number if provided', function () {
    var r = wrap(function () {
      return 53;
    });

    return r.then(function (v) {
      assert.strictEqual(v, 53);
    });
  });

  it('rejects with error if number is > 127', function () {
    var r = wrap(function () {
      return 123123;
    });

    return r.then(function () {
      assert.ok(false, 'should reject');
    }, function (err) {
      assert.ok(err);
      assert.ok(err.message.match(/Exit code must be a positive integer/));
    });
  });

  it('rejects with error if number < 0', function () {
    var r = wrap(function () {
      return -37;
    });

    return r.then(function () {
      assert.ok(false, 'should reject');
    }, function (err) {
      assert.ok(err);
      assert.ok(err.message.match(/Exit code must be a positive integer/));
    });
  });

  it('rejects wtih error if not a valid result', function () {
    var r = wrap(function () {
      return 'sdfsdf';
    });

    return r.then(function () {
      assert.ok(false, 'should reject');
    }, function (err) {
      assert.ok(err);
      assert.ok(err.message.match(/Exit code must be undefined, boolean/));
    });
  });
});
