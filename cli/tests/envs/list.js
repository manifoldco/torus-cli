/* eslint-env mocha */

'use strict';

var sinon = require('sinon');
var assert = require('assert');
var utils = require('common/utils');
var Promise = require('es6-promise').Promise;

var envs = require('../../lib/envs');
var client = require('../../lib/api/client').create();
var sessionMiddleware = require('../../lib/middleware/session');
var Config = require('../../lib/config');
var Context = require('../../lib/cli/context');
var Daemon = require('../../lib/daemon/object').Daemon;

var ORG = {
  id: utils.id('org'),
  body: {
    name: 'jeff-arigato-sh'
  }
};

var SERVICE = {
  id: utils.id('service'),
  body: {
    name: 'api-1',
    org_id: ORG.id
  }
};

var ENV = {
  id: utils.id('env'),
  body: {
    owner_id: SERVICE.id,
    service: SERVICE.body.name
  }
};

var ENV_PATH = '/envs';
var SERVICE_PATH = '/services';
var CTX_DAEMON_EMPTY;
var CTX;

describe('Envs List', function () {
  before(function () {
    this.sandbox = sinon.sandbox.create();
  });
  beforeEach(function () {
    this.sandbox.stub(envs.list.output, 'success');
    this.sandbox.stub(envs.list.output, 'failure');
    this.sandbox.stub(client, 'get')
      .returns(Promise.resolve({ body: [ENV] }));
    this.sandbox.spy(client, 'auth');

    // Context stub when no session set
    CTX_DAEMON_EMPTY = new Context({});
    CTX_DAEMON_EMPTY.config = new Config(process.cwd());
    CTX_DAEMON_EMPTY.daemon = new Daemon(CTX_DAEMON_EMPTY.config);

    // Context stub with session set
    CTX = new Context({});
    CTX.config = new Config(process.cwd());
    CTX.daemon = new Daemon(CTX.config);
    CTX.params = ['ABC123ABC'];

    // Empty daemon
    this.sandbox.stub(CTX_DAEMON_EMPTY.daemon, 'set')
      .returns(Promise.resolve());
    this.sandbox.stub(CTX_DAEMON_EMPTY.daemon, 'get')
      .returns(Promise.resolve({ token: '', passphrase: '' }));
    // Daemon with session
    this.sandbox.stub(CTX.daemon, 'set')
      .returns(Promise.resolve());
    this.sandbox.stub(CTX.daemon, 'get')
      .returns(Promise.resolve({ token: 'this is a token', passphrase: 'hi' }));
    // Run the session middleware to populate the context object
    return Promise.all([
      sessionMiddleware()(CTX),
      sessionMiddleware()(CTX_DAEMON_EMPTY)
    ]);
  });

  afterEach(function () {
    this.sandbox.restore();
  });

  describe('#execute', function () {
    it('authorizes the client', function () {
      return envs.list.execute(CTX).then(function () {
        sinon.assert.calledOnce(client.auth);
      }).catch(function () {
        assert.ok(false, 'should not error');
      });
    });

    it('sends an api request to envs', function () {
      client.get.onFirstCall().returns(Promise.resolve({ body: [SERVICE] }));
      client.get.onSecondCall().returns(Promise.resolve({ body: [ENV] }));

      return envs.list.execute(CTX).then(function (payload) {
        sinon.assert.calledWith(client.get.getCall(0), { url: SERVICE_PATH });
        sinon.assert.calledWith(client.get.getCall(1), {
          url: ENV_PATH,
          qs: { owner_id: SERVICE.id }
        });
        assert(payload, { services: [SERVICE], envs: [ENV] });
      }).catch(function () {
        assert.ok(false, 'did not expect to catch error');
      });
    });

    it('accepts optional [-s --service] flags', function () {
      CTX.options = { service: { value: SERVICE.body.name } };

      return envs.list.execute(CTX).then(function (payload) {
        sinon.assert.calledWith(client.get.getCall(0), {
          url: SERVICE_PATH + '/' + SERVICE.body.name
        });

        assert(payload, { services: [SERVICE], envs: [ENV] });
      }).catch(function () {
        assert.ok(false, 'did not expect to catch error');
      });
    });

    it('rejects invalid service names', function () {
      CTX.options = { service: { value: '~a~' } };
      var expErr = 'Only alphanumeric, hyphens and underscores are allowed';

      return envs.list.execute(CTX).then(function () {
        assert.ok(false, 'should not resolve');
      }).catch(function (err) {
        assert.ok(err);
        assert.strictEqual(err, expErr);
      });
    });
  });
});
