'use strict';

var assert = require('assert');
var sinon = require('sinon');

var Context = require('../../lib/cli/context');
var Option = require('../../lib/cli/option');
var Command = require('../../lib/cli/command');

describe('Command', function() {

  var c;
  describe('constructor', function() {
    
    it('throws error if usage is not a string', function() {
      assert.throws(function() {
        c = new Command(false);
      }, /Usage must be a string/);
    });

    it('throws error if usage is malformed', function() {
      assert.throws(function() {
        c = new Command('my@cmd');
      }, /Usage does not match regex/);
    });

    it('throws error if handler is not a function', function() {
      assert.throws(function() {
        c = new Command('my:cmd [usage]', false);
      }, /Handler must be a function or object with run method/);
    });

    it('throws error if handler is obj but no run method', function() {
      assert.throws(function() {
        var obj = { run: false };
        c = new Command('my:cmd', 'description', obj);
      });
    });

    it('sets group/subpath for root cmd', function() {
      function fn () {}

      c = new Command('signup <name> <password>', 'my description', fn);

      assert.strictEqual(c.slug, 'signup');
      assert.strictEqual(c.group, 'signup');
      assert.strictEqual(c.subpath, 'signup');
      assert.strictEqual(c._handler, fn);
      assert.strictEqual(c.description, 'my description');
    });

    it('sets group/subpath for sub cmd', function() {
      function fn () {}

      c = new Command('envs:create <name>', 'create env', fn);

      assert.strictEqual(c.slug, 'envs:create');
      assert.strictEqual(c.group, 'envs');
      assert.strictEqual(c.subpath, 'create');
      assert.strictEqual(c._handler, fn);
      assert.strictEqual(c.description, 'create env');
    });
  });

  describe('#option', function() {
    it('adds option if option obj provided', function() {
      c = new Command('envs', 'hello', function() {});

      c.option(new Option('-p, --pretty', 'pretty things'));

      assert.ok(c.options[0] instanceof Option);
      assert.strictEqual(c.options[0].name(), 'pretty');
    });

    it('adds option if c-tor params provided', function() {
      c = new Command('envs', 'hello', function() {});
      c.option('-p, --pretty', 'pretty things');

      assert.ok(c.options[0] instanceof Option);
      assert.strictEqual(c.options[0].name(), 'pretty');
    });
  });

  describe('#run', function() {
    
    it('runs the pre/post middleware', function() {
      var preSpy = sinon.spy();
      var postSpy = sinon.spy();
      var ctx = new Context({});

      c = new Command('envs', 'do something', function() {});
      c.hook('pre', preSpy);
      c.hook('post', postSpy);

      return c.run(ctx).then(function() {
        sinon.assert.calledOnce(preSpy);
        sinon.assert.calledWith(postSpy, ctx);
        sinon.assert.calledOnce(postSpy);
        sinon.assert.calledWith(postSpy, ctx);
      });
    });

    it('runs a function', function() {
      var spy = sinon.spy();
      var ctx = new Context({});

      c = new Command('envs', 'do something', spy);

      return c.run(ctx).then(function() {
        sinon.assert.calledOnce(spy);
        sinon.assert.calledWith(spy, ctx);
      });
    });

    it('runs a obj#run', function() {
      var o = { run: sinon.spy() };
      var ctx = new Context({});

      c = new Command('envs', 'do something', o);

      return c.run(ctx).then(function() {
        sinon.assert.calledOnce(o.run);
        sinon.assert.calledWith(o.run, ctx);
      });
    })
  });
});
