'use strict';

const assert = require('assert');
const explore = require('../../lib/util/explore');

describe('explore', function() {
  it('finds ref inside a simple object', function() {
    var data = {
      a: { '$ref': 'schema.json#' }
    };

    assert.deepEqual(explore(data), [
      { objPath: '.a', filePath: 'schema.json#' }
    ]);
  });

  it('finds refs inside a nested object', function() {
    var data = {
      a: {
        b: { '$ref': 'schema.json#' }
      }
    };

    assert.deepEqual(explore(data), [
      { objPath: '.a.b', filePath: 'schema.json#' }
    ]);
  });

  it('finds refs inside an array', function() {
    var data = {
      a: {
        b: [ 'a', { '$ref': 'schema.json#' } ]
      },
      c: { '$ref': 'schema.json#' }
    };

    assert.deepEqual(explore(data), [
      { objPath: '.a.b[1]', filePath: 'schema.json#' },
      { objPath: '.c', filePath: 'schema.json#' }
    ]);
  });

  it('errors if it doesnt get an array or object', function() {
    assert.throws(function () { explore('a'); }, TypeError);
  });
});
