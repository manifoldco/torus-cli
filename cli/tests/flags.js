/* eslint-env mocha */

'use strict';

var assert = require('assert');

var flags = require('../lib/flags');
var Command = require('../lib/cli/command');

describe('flags', function () {
  var cmd;
  beforeEach(function () {
    cmd = new Command('test', 'testing', function () {});
  });

  describe('add', function () {
    it('throws if a Command is not passed', function () {
      assert.throws(function () {
        flags.add(null);
      }, /Must provide an instance of a Command/);
    });

    it('throws if the option is unknown', function () {
      assert.throws(function () {
        flags.add(cmd, 'sdfs');
      }, /Unknown option: sdfs/);
    });

    it('throws if the option has already been added', function () {
      assert.throws(function () {
        flags.add(cmd, 'project');
        flags.add(cmd, 'project');
      }, /Cannot add the same option twice/);
    });

    it('sets the option on the command', function () {
      flags.add(cmd, 'project');

      assert.strictEqual(cmd.options.length, 1);
      assert.strictEqual(cmd.options[0].name(), 'project');
    });

    it('sets the option on the command with overrides', function () {
      flags.add(cmd, 'project', {
        description: 'test',
        default: 'woo'
      });

      assert.strictEqual(cmd.options.length, 1);
      assert.strictEqual(cmd.options[0].name(), 'project');
      assert.strictEqual(cmd.options[0].description, 'test');
      assert.strictEqual(cmd.options[0].defaultValue, 'woo');
    });
  });
});
