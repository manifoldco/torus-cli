/* eslint-env mocha */

'use strict';

var assert = require('assert');

var Target = require('../../lib/context/target');

describe('Target', function () {
  var target;
  describe('constructor', function () {
    it('throws error if path is not a string', function () {
      assert.throws(function () {
        target = new Target(false);
      }, /Must provide a path string/);
    });

    it('throws error if path is not absolute', function () {
      assert.throws(function () {
        target = new Target('bcdefgh/g');
      }, /Must provide an absolute path/);
    });

    it('throws error if context object is not provided', function () {
      assert.throws(function () {
        target = new Target('/a/b/c', false);
      }, /Must provide context object/);
    });

    it('assigns the properties correctly', function () {
      target = new Target('/a/b/c', {
        org: 'a',
        project: 'b',
        environment: 'c',
        service: 'd'
      });

      assert.strictEqual(target.org, 'a');
      assert.strictEqual(target.project, 'b');
      assert.strictEqual(target.environment, 'c');
      assert.strictEqual(target.service, 'd');
    });
  });

  describe('flags', function () {
    beforeEach(function () {
      target = new Target('/a/b/c/d', {
        org: 'a',
        project: 'b',
        environment: 'c',
        service: 'd'
      });
    });

    it('assigns the properties', function () {
      target.flags({
        org: 'd',
        project: 'c',
        environment: 'e',
        service: 'f'
      });

      assert.strictEqual(target.org, 'd');
      assert.strictEqual(target.project, 'c');
      assert.strictEqual(target.environment, 'e');
      assert.strictEqual(target.service, 'f');
    });

    it('does not assign undefined values', function () {
      target.flags({
        service: undefined
      });

      assert.strictEqual(target.org, 'a');
      assert.strictEqual(target.service, 'd');
    });

    it('resets lower-level values if override using flags - org', function () {
      target.flags({
        org: 'joe'
      });

      assert.strictEqual(target.org, 'joe');
      assert.strictEqual(target.project, null);
      assert.strictEqual(target.environment, null);
      assert.strictEqual(target.service, null);
    });

    it('resets lower-level values if override using flags - proj', function () {
      target.flags({
        project: 'woo'
      });

      assert.strictEqual(target.org, 'a');
      assert.strictEqual(target.project, 'woo');
      assert.strictEqual(target.environment, null);
      assert.strictEqual(target.service, null);
    });

    it('doesnt reset if service is different', function () {
      target.flags({
        service: 'jack'
      });

      assert.strictEqual(target.org, 'a');
      assert.strictEqual(target.project, 'b');
      assert.strictEqual(target.service, 'jack');
      assert.strictEqual(target.environment, 'c');
    });
  });
});
