'use strict';

var assert = require('assert');

var deepPick = require('../../lib/util/deep_pick');

describe('deepPick', function() {

  it('handles single level', function() {
    var obj = {
      a: 1
    };

    assert.strictEqual(deepPick('a', obj), 1);
  });

  it('handles even multiple depth', function() {
    var obj = {
      a: {
        b: 2
      }
    };

    assert.strictEqual(deepPick('a.b', obj), 2);
  });

  it('handles odd depth', function() {
    var obj = {
      a: {
        b: {
          c: 3
        }
      }
    };

    assert.strictEqual(deepPick('a.b.c', obj), 3);
  });

  it('errors if its a non object', function() {
    var obj = {
      a: 1
    };

    assert.throws(function() {
      deepPick('a.b', obj);
    }, TypeError);
  });

  it('returns undefined if it doesnt exist', function() {
    var obj = { a: { b: { c: 4 } } };

    assert.strictEqual(deepPick('a.b.d', obj), undefined);
  });

  it('sets as null', function() {
    var obj = {a : { b: null } };

    assert.strictEqual(deepPick('a.b', obj), null);
  });
});
