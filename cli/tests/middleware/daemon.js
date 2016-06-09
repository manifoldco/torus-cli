/* eslint-env mocha */

'use strict';

var assert = require('assert');

var sinon = require('sinon');
var Promise = require('es6-promise').Promise;

var Context = require('../../lib/cli/context');
var Config = require('../../lib/config');
var Daemon = require('../../lib/daemon/object').Daemon;
var daemon = require('../../lib/daemon');
var middleware = require('../../lib/middleware/daemon');

describe('daemon middleware', function () {
  var sandbox;
  var ctx;

  beforeEach(function () {
    sandbox = sinon.sandbox.create();
    ctx = new Context({});
    ctx.config = new Config(process.cwd());
  });
  afterEach(function () {
    sandbox.restore();
  });

  describe('preHook', function () {
    it('retrieves the daemon', function () {
      var d = new Daemon(ctx.config);
      sandbox.stub(daemon, 'get').returns(Promise.resolve(d));

      return middleware.preHook()(ctx).then(function () {
        assert.ok(ctx.daemon instanceof Daemon);
        assert.strictEqual(ctx.daemon, d);
      });
    });

    it('starts daemon if its not running', function () {
      var d = new Daemon(ctx.config);

      sandbox.stub(daemon, 'get').returns(Promise.resolve(null));
      sandbox.stub(daemon, 'start').returns(Promise.resolve(d));

      return middleware.preHook()(ctx).then(function () {
        assert.ok(ctx.daemon instanceof Daemon);
        assert.strictEqual(ctx.daemon, d);
      });
    });
  });

  describe('postHook', function () {
    it('disconnects from the running daemon', function () {
      var d = new Daemon(ctx.config);
      ctx.daemon = d;

      sandbox.stub(d, 'connected').returns(true);
      sandbox.stub(d, 'disconnect').returns(Promise.resolve());

      return middleware.postHook()(ctx).then(function () {
        sinon.assert.calledOnce(d.disconnect);
      });
    });
  });
});
