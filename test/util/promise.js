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

  describe('.series', function() {
  
    it('executes in order', function() {
      var x = 0;
      function fn () {
        return new Promise((resolve) => {
          setTimeout(function() {
            x += 1;
            resolve(x-1);
          }, Math.random()*25);
        });
      }

      return promise.series([fn, fn, fn]).then((results) => {
        assert.deepEqual([0,1,2], results);
      });
    });

    it('immediately returns an error', function() {
      var x = 0;
      function fn () {
        return new Promise((resolve, reject) => {
          setTimeout(function() {
            x += 1;

            if (x ===  2) {
              return reject(new Error('woo'));
            } else {
              return resolve(x-1);
            }
          }, Math.random()*25);
        });
      }

      return promise.series([fn, fn, fn]).then(() => {
        assert.ok(false, 'this shoudlnt be called');
      }, (err) => {
        assert.ok(err instanceof Error);
        assert.strictEqual(err.message, 'woo');
      });

    });
  });

  describe('.seriesMap', function() {
    it('maintains object.keys order', function() {
      var x = 0;
      function fn () {
        return new Promise((resolve) => {
          x = x + 1;
          resolve(x);
        });
      }

      return promise.seriesMap({ a: fn, b: fn, c: fn }).then((results) => {
        assert.deepEqual({ a: 1, b: 2, c: 3 }, results);
      });
    });

    it('returns on an error', function () {
      var x = 0;
      function fn () {
        return new Promise((resolve, reject) => {
          x = x + 1;
          if (x == 2) {
            return reject(new Error('hi'));
          } else {
            return resolve(x);
          }
        });
      }

      return promise.seriesMap({ a: fn, b: fn, c: fn }).then(() => {
        assert.ok(false, 'this shoudlnt happen');
      }, (err) => {
        assert.ok(err instanceof Error);
        assert.strictEqual(err.message, 'hi');
      });
    });
  });
});
