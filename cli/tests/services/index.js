/* eslint-env mocha */

'use strict';

var sinon = require('sinon');
var assert = require('assert');
var utils = require('common/utils');
var Promise = require('es6-promise').Promise;

var services = require('../../lib/services/create');
var client = require('../../lib/api/client').create();
var Config = require('../../lib/config');
var Context = require('../../lib/cli/context');
var Daemon = require('../../lib/daemon/object').Daemon;
var sessionMiddleware = require('../../lib/middleware/session');

var ORG = {
  id: utils.id('org'),
  body: {
    name: 'my-org'
  }
};

var SERVICE = {
  id: utils.id('service'),
  body: {
    name: 'api-1',
    org_id: ORG.id
  }
};

var CTX_DAEMON_EMPTY;
var CTX;

describe('Services Create', function () {
  before(function () {
    this.sandbox = sinon.sandbox.create();
  });
  beforeEach(function () {
    this.sandbox.stub(services.output, 'success');
    this.sandbox.stub(services.output, 'failure');
    this.sandbox.stub(client, 'post')
      .returns(Promise.resolve(SERVICE));
    this.sandbox.spy(client, 'auth');

    // Context stub when no token set
    CTX_DAEMON_EMPTY = new Context({});
    CTX_DAEMON_EMPTY.config = new Config(process.cwd());
    CTX_DAEMON_EMPTY.daemon = new Daemon(CTX_DAEMON_EMPTY.config);

    // Context stub with token set
    CTX = new Context({});
    CTX.config = new Config(process.cwd());
    CTX.daemon = new Daemon(CTX.config);
    CTX.params = ['ABC123ABC'];

    // Empty daemon
    this.sandbox.stub(CTX_DAEMON_EMPTY.daemon, 'set')
      .returns(Promise.resolve());
    this.sandbox.stub(CTX_DAEMON_EMPTY.daemon, 'get')
      .returns(Promise.resolve({ token: '', passphrase: '' }));
    // Daemon with token
    this.sandbox.stub(CTX.daemon, 'set')
      .returns(Promise.resolve());
    this.sandbox.stub(CTX.daemon, 'get')
      .returns(Promise.resolve({
        token: 'this is a token',
        passphrase: 'a passphrase'
      }));
    // Run the token middleware to populate the context object
    return Promise.all([
      sessionMiddleware()(CTX),
      sessionMiddleware()(CTX_DAEMON_EMPTY)
    ]);
  });
  afterEach(function () {
    this.sandbox.restore();
  });
  describe('execute', function () {
    it('calls _execute with inputs', function () {
      this.sandbox.stub(services, '_prompt').returns(Promise.resolve());
      this.sandbox.stub(services, '_execute').returns(Promise.resolve());
      return services.execute(CTX).then(function () {
        sinon.assert.calledOnce(services._execute);
      });
    });
    it('skips the prompt when inputs are supplied', function () {
      this.sandbox.stub(services, '_prompt').returns(Promise.resolve());
      this.sandbox.stub(services, '_execute').returns(Promise.resolve());
      return services.execute(CTX).then(function () {
        sinon.assert.notCalled(services._prompt);
      });
    });
  });
  describe('_execute', function () {
    it('authorizes the client', function () {
      var input = { name: 'api-1' };
      return services._execute(CTX.session, input).then(function () {
        sinon.assert.calledOnce(client.auth);
      });
    });

    it('errors if session is missing', function () {
      var session = CTX_DAEMON_EMPTY.session;
      var input = {};
      return services._execute(session, input).then(function () {
        assert.ok(false, 'should error');
      }).catch(function (err) {
        assert.ok(err);
        assert.strictEqual(err.message, 'Session object missing on Context');
      });
    });

    it('sends api request to services', function () {
      var input = { name: 'api-1' };
      return services._execute(CTX.session, input).then(function () {
        sinon.assert.calledOnce(client.post);
        var firstCall = client.post.firstCall;
        var args = firstCall.args;
        assert.deepEqual(args[0], {
          url: '/services',
          json: {
            body: {
              name: 'api-1'
            }
          }
        });
      });
    });
  });
});
