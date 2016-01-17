'use strict';

var assert = require('assert');
var deepSet = require('../../lib/util/deep_set');

describe('deepSet', function() {

  it('handles level zero depth set', function() {
    assert.deepEqual(deepSet('a', 1, {}), { a: 1 });
  });

  it('handles multiple set depth', function() {
    assert.deepEqual(deepSet('a.b.c', 2, { a: { b: { d: 1 } } } ), {
      a: {
        b: {
          d: 1,
          c: 2
        }
      }
    });
  });

  it('errors if non object provided', function() {
    assert.throws(function() {
      deepSet('a.b', 1, { a: 3 });
    }, TypeError);
  });

  it('expands with objects if it doesnt exist', function () {
  
    assert.deepEqual(deepSet('a.b.c.d.e', 5, { a: { c: 1 } }), {
      a: {
        c: 1,
        b: {
          c: {
            d: {
              e: 5
            }
          }
        }
      }
    });
  });
});
