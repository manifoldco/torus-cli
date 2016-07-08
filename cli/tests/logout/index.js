/* eslint-env mocha */

'use strict';

var assert = require('assert');
var sinon = require('sinon');
var Promise = require('es6-promise').Promise;

var logout = require('../../lib/logout');

var Session = require('../../lib/session');
var Config = require('../../lib/config');
var Context = require('../../lib/cli/context');
var Daemon = require('../../lib/daemon/object').Daemon;
var api = require('../../lib/api');

describe('Logout', function () {
  var ctx;
  before(function () {
    this.sandbox = sinon.sandbox.create();
  });

  beforeEach(function () {
    ctx = new Context({});

    ctx.config = new Config(process.cwd());
    ctx.daemon = new Daemon(ctx.config);
    ctx.session = new Session({
      token: 'aedfasdf',
      passphrase: 'asdfsadf'
    });
    ctx.api = api.build({ auth_token: ctx.session.token });

    this.sandbox.spy(ctx.api, 'reset');
    this.sandbox.stub(ctx.api.tokens, 'remove').returns(Promise.resolve());
    this.sandbox.stub(ctx.daemon, 'logout').returns(Promise.resolve());
  });

  afterEach(function () {
    this.sandbox.restore();
  });

  it('sends a logout request to the registry and daemon', function () {
    ctx.daemon.logout.returns(Promise.resolve());

    return logout(ctx).then(function () {
      sinon.assert.calledOnce(ctx.daemon.logout);
      sinon.assert.calledOnce(ctx.api.tokens.remove);
      sinon.assert.calledOnce(ctx.api.reset);

      assert.strictEqual(ctx.session, null);
    });
  });

  it('reports an error and destroys the token on client failure', function () {
    ctx.api.tokens.remove.returns(Promise.reject(new Error('hi')));
    ctx.daemon.logout.returns(Promise.resolve());

    return logout(ctx).catch(function () {
      sinon.assert.calledOnce(ctx.daemon.logout);
      sinon.assert.calledOnce(ctx.api.tokens.remove);
      sinon.assert.calledOnce(ctx.api.reset);

      assert.strictEqual(ctx.session, null);
    });
  });

  it('reports an error and destroys the token on daemon failure', function () {
    ctx.daemon.logout.returns(Promise.reject(new Error('err')));
    return logout(ctx).catch(function () {
      sinon.assert.calledOnce(ctx.daemon.logout);
      sinon.assert.calledOnce(ctx.api.tokens.remove);
      sinon.assert.calledOnce(ctx.api.reset);

      assert.strictEqual(ctx.session, null);
    });
  });

  it('errors if session object is not available', function () {
    ctx.session = null;
    return logout(ctx).then(function () {
      assert.ok(false, 'should error');
    }).catch(function (err) {
      assert.ok(err instanceof Error);
      assert.strictEqual(err.message, 'Session object not on Context object');
    });
  });
});
