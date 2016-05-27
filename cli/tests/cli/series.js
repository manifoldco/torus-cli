/* eslint-env mocha */

'use strict';

var assert = require('assert');
var Promise = require('es6-promise').Promise;

var series = require('../../lib/cli/series');

describe('.series', function () {
  it('executes in order', function () {
    var x = 0;
    function fn() {
      return new Promise(function (resolve) {
        setTimeout(function () {
          x += 1;
          resolve(x - 1);
        }, Math.random() * 25);
      });
    }

    return series([fn, fn, fn]).then(function (results) {
      assert.deepEqual([0, 1, 2], results);
    });
  });

  it('immediately returns an error', function () {
    var x = 0;
    function fn() {
      return new Promise(function (resolve, reject) {
        setTimeout(function () {
          x += 1;

          if (x === 2) {
            return reject(new Error('woo'));
          }

          return resolve(x - 1);
        }, Math.random() * 25);
      });
    }

    return series([fn, fn, fn]).then(function () {
      assert.ok(false, 'this shoudlnt be called');
    }, function (err) {
      assert.ok(err instanceof Error);
      assert.strictEqual(err.message, 'woo');
    });
  });
});
