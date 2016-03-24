'use strict';

var assert = require('assert');

var Context =  require('../../lib/cli/context');
var Option = require('../../lib/cli/option');

describe('Option', function() {
  
  describe('constructor', function() {
    it('sets required if required', function() {
      var o = new Option('-p, --pretty <name>', 'description');

      assert.strictEqual(o.required, true, 'required');
      assert.strictEqual(o.optional, false, 'optional');
      assert.strictEqual(o.hasParam, true, 'hasParam');
      assert.strictEqual(o.bool, false, 'bool');
      assert.strictEqual(o.description, 'description');

      assert.strictEqual(o.name(), 'pretty');
      assert.strictEqual(o.shortcut(), 'p');
    });

    it('sets optional if optioned', function() {
      var o = new Option('-p, --pretty [name]');

      assert.strictEqual(o.required, false);
      assert.strictEqual(o.optional, true);
      assert.strictEqual(o.hasParam, true);
      assert.strictEqual(o.bool, false);

      assert.strictEqual(o.name(), 'pretty');
      assert.strictEqual(o.shortcut(), 'p');
    });

    it('sets both optional and required', function() {
      var o = new Option('-p, --pretty <name> [longname]');

      assert.strictEqual(o.required, true);
      assert.strictEqual(o.optional, true);
      assert.strictEqual(o.hasParam, true);
      assert.strictEqual(o.bool, false);

      assert.strictEqual(o.name(), 'pretty');
      assert.strictEqual(o.shortcut(), 'p');
    });

    it('errors if optional are before required', function() {
      assert.throws(function() {
        var o = new Option('-p, --pretty [name] <longname>');
        /*jshint unused: true*/
      }, /Required must come before/);
    });

    it('handles boolean', function () {
      var o = new Option('-s, --save', 'our description');
      
      assert.strictEqual(o.hasParam, false, 'hasParam');
      assert.strictEqual(o.bool, true, 'bool');
      assert.strictEqual(o.defaultValue, false);
    });

    it('handles boolean with --no-*', function() {
      var o = new Option('-n, --no-save', 'no saving');

      assert.strictEqual(o.hasParam, false);
      assert.strictEqual(o.bool, true, 'bool');
      assert.strictEqual(o.defaultValue, true, 'defaultValue');
    });
  });

  describe('#evaluate', function() {

    var c;
    beforeEach(function() {
      c = new Context({});
    })

    it('sets the value', function() {
      var o = new Option('-n, --name <name>', 'set name');
      
      o.evaluate(c, {name: 'hi'});

      assert.strictEqual(o.value, 'hi');
      assert.strictEqual(c.prop('name'), o);
    });

    it('sets the value with shortcut', function() {
      var o = new Option('-n, --name <name>', 'set name');

      o.evaluate(c, {n: 'hi'});
      assert.strictEqual(o.value, 'hi');
      assert.strictEqual(c.prop('name'), o);
    });

    it('sets the default value', function() {
      var o = new Option('-n, --name [name]', 'set name', 'joe');

      o.evaluate(c, {});

      assert.strictEqual(o.value, 'joe');
      assert.strictEqual(c.prop('name'), o);
    });

    it('sets as undefined if not bool', function() {
      var o = new Option('n, --name [name]', 'set name');

      o.evaluate(c, {});

      assert.strictEqual(o.value, undefined);
      assert.strictEqual(c.prop('name'), o);
    });

    it('sets as false if bool', function() {
      var o = new Option('-n, --no', 'stuff');

      o.evaluate(c, {});

      assert.strictEqual(o.value, false);
      assert.strictEqual(c.prop('no'), o);
    });

    it('sets as true if no-bool', function() {
      var o = new Option('-x, --no-x', 'stuff');

      o.evaluate(c, {});
      assert.strictEqual(o.value, true);
      assert.strictEqual(c.prop('x'), o);
    });
  });
});
