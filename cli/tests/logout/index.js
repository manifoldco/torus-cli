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
    this.sandbox.stub(ctx.api.logout, 'post').returns(Promise.resolve());
  });

  afterEach(function () {
    this.sandbox.restore();
  });

  it('sends a logout request to the registry and daemon', function () {
    return logout(ctx).then(function () {
      sinon.assert.calledOnce(ctx.api.logout.post);
      sinon.assert.calledOnce(ctx.api.reset);

      assert.strictEqual(ctx.session, null);
    });
  });

  it('reports an error on client failure', function () {
    ctx.api.logout.post.returns(Promise.reject(new Error('hi')));

    return logout(ctx).catch(function () {
      sinon.assert.calledOnce(ctx.api.logout.post);
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
