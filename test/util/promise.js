'use strict';

const assert = require('assert');
const promise = require('../../lib/util/promise');

describe('promise util', function() {
  describe('.map', function() {
    
    it('errors if its not an object', function() {
      return promise.map('a').then(() => {
        assert.ok(false, 'did not error');
      }, (err) => {
        assert.ok(err instanceof TypeError);
        assert.strictEqual(err.message, 'Must be a plain object');
      });
    });

    it('maps promises to resolved values', function() {
      var map = {
        a: Promise.resolve('a'),
        b: Promise.resolve('b')
      };
    
      return promise.map(map).then((map) => {
        assert.deepEqual(map, {
          a: 'a',
          b: 'b'
        });
      });
    });

    it('rejects if one promise is rejected', function() {
      var map = {
        a: Promise.resolve('a'),
        b: Promise.reject(new Error('no'))
      };

      return promise.map(map).then(() => {
        assert.ok(false, 'shouldnt have succeeded');
      }, (err) => {
        assert.ok(err instanceof Error);
        assert.strictEqual(err.message, 'no');
      });
    });
  });
});
