'use strict';

var assert = require('assert');
var Promise = require('es6-promise').Promise;

var series = require('../../lib/cli/series');

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

    return series([fn, fn, fn]).then((results) => {
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

    return series([fn, fn, fn]).then(() => {
      assert.ok(false, 'this shoudlnt be called');
    }, (err) => {
      assert.ok(err instanceof Error);
      assert.strictEqual(err.message, 'woo');
    });

  });
});
