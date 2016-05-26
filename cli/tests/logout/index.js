/* eslint-env mocha */

'use strict';

var assert = require('assert');
var sinon = require('sinon');
var Promise = require('es6-promise').Promise;

var logout = require('../../lib/logout');
var client = require('../../lib/api/client').create();

var Session = require('../../lib/session');
var Config = require('../../lib/config');
var Context = require('../../lib/cli/context');
var Daemon = require('../../lib/daemon/object').Daemon;


describe('Logout', function () {
  var ctx;
  var session;
  before(function () {
    this.sandbox = sinon.sandbox.create();
  });

  beforeEach(function () {
    ctx = new Context({});

    ctx.config = new Config(process.cwd());
    ctx.daemon = new Daemon(ctx.config);
    session = ctx.session = new Session({
      token: 'aedfasdf',
      passphrase: 'asdfsadf'
    });

    this.sandbox.stub(ctx.daemon, 'logout');
    this.sandbox.stub(client, 'post');
    this.sandbox.stub(client, 'get');
    this.sandbox.stub(client, 'delete');
    this.sandbox.stub(client, 'auth');
    this.sandbox.stub(client, 'reset');
  });

  afterEach(function () {
    this.sandbox.restore();
  });

  it('sends a logout request to the registry and daemon', function () {
    client.post.returns(Promise.resolve());
    ctx.daemon.logout.returns(Promise.resolve());

    return logout(ctx).then(function () {
      sinon.assert.calledWith(client.auth, session.token);
      sinon.assert.calledWith(client.delete,
                              { url: '/session/' + session.token });
      sinon.assert.calledOnce(ctx.daemon.logout);
      sinon.assert.calledOnce(client.reset);
      assert.strictEqual(ctx.session, null);
    });
  });

  it('reports an error and destroys the token on client failure', function () {
    client.post.returns(Promise.reject(new Error('err')));
    ctx.daemon.logout.returns(Promise.resolve());

    return logout(ctx).catch(function () {
      sinon.assert.calledWith(client.auth, session.token);
      sinon.assert.calledOnce(ctx.daemon.logout);
      sinon.assert.calledOnce(client.reset);
      assert.strictEqual(ctx.session, null);
    });
  });

  it('reports an error and destroys the token on daemon failure', function () {
    client.post.returns(Promise.resolve());
    ctx.daemon.logout.returns(Promise.reject(new Error('err')));

    return logout(ctx).catch(function () {
      sinon.assert.calledWith(client.auth, session.token);
      sinon.assert.calledWith(client.delete,
                              { url: '/session/' + session.token });
      sinon.assert.calledOnce(client.reset);
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
