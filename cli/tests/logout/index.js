'use strict';

var sinon = require('sinon');
var Promise = require('es6-promise').Promise;

var logout = require('../../lib/logout');
var client = require('../../lib/api/client').create();

var Config = require('../../lib/config');
var Context = require('../../lib/cli/context');
var Daemon = require('../../lib/daemon/object').Daemon;

var CTX = new Context({});

CTX.config = new Config(process.cwd());
CTX.daemon = new Daemon(CTX.config);
CTX.token = 'aToken';

describe('Logout', function() {

  before(function() {
    this.sandbox = sinon.sandbox.create();
  });

  beforeEach(function() {
    this.sandbox.stub(CTX.daemon, 'logout');
    this.sandbox.stub(client, 'post');
    this.sandbox.stub(client, 'auth');
    this.sandbox.stub(client, 'reset');
  });

  afterEach(function() {
    this.sandbox.restore();
  });

  it('sends a logout request to the registry and daemon', function() {
    client.post.returns(Promise.resolve());
    CTX.daemon.logout.returns(Promise.resolve());

    return logout(CTX).then(function() {
      sinon.assert.calledWith(client.auth, CTX.token);
      sinon.assert.calledWith(client.post, { url: '/logout' });
      sinon.assert.calledOnce(CTX.daemon.logout);
      sinon.assert.calledOnce(client.reset);
    });
  });

  it('reports an error and destroys the token on client failure', function() {
    client.post.returns(Promise.reject(new Error('err')));
    CTX.daemon.logout.returns(Promise.resolve());

    return logout(CTX).catch(function() {
      sinon.assert.calledWith(client.auth, CTX.token);

      sinon.assert.calledOnce(CTX.daemon.logout);
      sinon.assert.calledOnce(client.reset);
    });
  });

  it('reports an error and destroys the token on daemon failure', function() {
    client.post.returns(Promise.resolve());
    CTX.daemon.logout.returns(Promise.reject(new Error('err')));

    return logout(CTX).catch(function() {
      sinon.assert.calledWith(client.auth, CTX.token);

      sinon.assert.calledWith(client.post, { url: '/logout' });
      sinon.assert.calledOnce(client.reset);
    });
  });
});
