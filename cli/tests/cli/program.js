/* eslint-env mocha */

'use strict';

var assert = require('assert');
var sinon = require('sinon');
var Promise = require('es6-promise').Promise;

var Program = require('../../lib/cli/program');
var Command = require('../../lib/cli/command');

describe('Program', function () {
  var sandbox;
  var p;
  var c;

  beforeEach(function () {
    sandbox = sinon.sandbox.create();
  });

  afterEach(function () {
    sandbox.restore();
  });

  describe('constructor', function () {
    it('errors on name', function () {
      assert.throws(function () {
        p = new Program(false);
      }, /A string must be provided for the name/);
    });

    it('errors on version', function () {
      assert.throws(function () {
        p = new Program('ag', 1);
      }, /A version must be provided/);
    });

    it('errors if template is missing', function () {
      assert.throws(function () {
        p = new Program('ag', '1.0.0', null);
      }, /Templates must be defined/);
    });

    it('constructs', function () {
      p = new Program('ag', '1.2.1', {
        program: 'abc',
        command: '123'
      });
    });
  });

  describe('#command', function () {
    beforeEach(function () {
      p = new Program('ag', '1.0.0', {
        program: 'hi',
        command: 'boo'
      });
    });

    it('errors if cmd is not a Command', function () {
      assert.throws(function () {
        p.command(false);
      }, /A Command object must be provided/);
    });

    it('sets group properly for root cmd', function () {
      c = new Command('hi', 'hello', function () {});
      p.command(c);

      assert.strictEqual(c.program, p);
      assert.strictEqual(p.commands[c.slug], c);
      assert.strictEqual(p.groups[c.group][c.subpath], c);
    });

    it('sets group properly for sub cmd', function () {
      c = new Command('hi:hello', 'hello', function () {});
      var c2 = new Command('hi:bye', 'bye', function () {});
      p.command(c);
      p.command(c2);


      assert.strictEqual(c.program, p);
      assert.strictEqual(p.commands[c.slug], c);
      assert.strictEqual(p.groups[c.group][c.subpath], c);
      assert.strictEqual(p.commands[c2.slug], c2);
      assert.strictEqual(p.groups[c2.group][c2.subpath], c2);
    });
  });

  describe('#run', function () {
    beforeEach(function () {
      sandbox = sinon.sandbox.create();

      p = new Program('ag', '1.2.1', {
        command: 'a',
        program: 'b'
      });
      p.command(new Command('hi', 'hello', function () {}));
      p.command(new Command('hi:bye', 'bye', function () {}));
    });

    afterEach(function () {
      sandbox.restore();
    });

    it('handles empty params', function () {
      sandbox.stub(p, '_rootHelp', function () {
        return Promise.resolve();
      });

      return p.run(['x', 'y']).then(function () {
        sinon.assert.calledOnce(p._rootHelp);
      });
    });

    it('handles help cmd with no args', function () {
      sandbox.stub(p, '_rootHelp', function () {
        return Promise.resolve();
      });

      return p.run(['x', 'y', 'help']).then(function () {
        sinon.assert.calledOnce(p._rootHelp);
      });
    });

    it('hanldes help cmd with bad args', function () {
      sandbox.stub(p, '_rootHelp', function () {
        return Promise.resolve();
      });
      sandbox.stub(console, 'log');

      return p.run(['x', 'y', 'help', 'g2']).then(function () {
        sinon.assert.calledOnce(console.log);
        sinon.assert.calledWith(console.log, 'Unknown command: g2');
        sinon.assert.calledOnce(p._rootHelp);
      });
    });

    it('handles help cmd with valid args', function () {
      sandbox.stub(p, '_cmdHelp', function () {
        return Promise.resolve();
      });

      return p.run(['x', 'y', 'help', 'hi']).then(function () {
        sinon.assert.calledOnce(p._cmdHelp);
      });
    });

    it('handles valid cmd', function () {
      sandbox.stub(p.commands.hi, 'run', function () {
        return Promise.resolve(true);
      });

      return p.run(['x', 'y', 'hi']).then(function () {
        sinon.assert.calledOnce(p.commands.hi.run);
      });
    });

    it('handles invalid cmd', function () {
      sandbox.stub(p, '_rootHelp', function () {
        return Promise.resolve();
      });
      sandbox.stub(console, 'log');

      return p.run(['x', 'y', 'g2']).then(function () {
        sinon.assert.calledOnce(p._rootHelp);
        sinon.assert.calledOnce(console.log);
        sinon.assert.calledWith(console.log, 'Unknown Command: g2');
      });
    });

    it('runs pre/post hooks', function () {
      var preSpy = sinon.spy();
      var postSpy = sinon.spy();

      p.hook('pre', preSpy);
      p.hook('post', postSpy);

      return p.run(['x', 'x', 'hi']).then(function () {
        sinon.assert.calledOnce(preSpy);
        sinon.assert.calledOnce(postSpy);
      });
    });

    it('resolves success as false if pre middleware fails', function () {
      var failStub = sinon.stub().returns(Promise.resolve(false));
      var passStub = sinon.stub().returns(Promise.resolve(false));

      p.hook('pre', failStub);
      p.hook('pre', passStub);

      return p.run(['x', 'x', 'hi']).then(function (success) {
        assert.strictEqual(success, false);

        sinon.assert.calledOnce(failStub);
        sinon.assert.notCalled(passStub);
      });
    });

    it('resolves success as false if post middleware fails', function () {
      var failStub = sinon.stub().returns(Promise.resolve(false));
      var passStub = sinon.stub().returns(Promise.resolve(false));

      p.hook('post', failStub);
      p.hook('post', passStub);

      return p.run(['x', 'x', 'hi']).then(function (success) {
        assert.strictEqual(success, false);

        sinon.assert.calledOnce(failStub);
        sinon.assert.notCalled(passStub);
      });
    });
  });
});
