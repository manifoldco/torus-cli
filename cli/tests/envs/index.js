/* eslint-env mocha */
'use strict';

var sinon = require('sinon');
var assert = require('assert');
var utils = require('common/utils');
var Promise = require('es6-promise').Promise;

var envs = require('../../lib/envs/create');
var client = require('../../lib/api/client').create();
var tokenMiddleware = require('../../lib/middleware/token');
var Config = require('../../lib/config');
var Context = require('../../lib/cli/context');
var Daemon = require('../../lib/daemon/object').Daemon;
var ValidationError = require('../../lib/validate').ValidationError;

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
    name: 'staging',
    owner_id: SERVICE.id
  }
};

var CTX_DAEMON_EMPTY;
var CTX;

describe('Envs Create', function () {
  before(function () {
    this.sandbox = sinon.sandbox.create();
  });
  beforeEach(function () {
    this.sandbox.stub(envs.output, 'success');
    this.sandbox.stub(envs.output, 'failure');
    this.sandbox.stub(client, 'get')
      .returns(Promise.resolve({ body: [SERVICE] }));
    this.sandbox.stub(client, 'post')
      .returns(Promise.resolve({ body: [ENV] }));
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
      .returns(Promise.resolve({ token: '' }));
    // Daemon with token
    this.sandbox.stub(CTX.daemon, 'set')
      .returns(Promise.resolve());
    this.sandbox.stub(CTX.daemon, 'get')
      .returns(Promise.resolve({ token: 'this is a token' }));
    // Run the token middleware to populate the context object
    return Promise.all([
      tokenMiddleware.preHook()(CTX),
      tokenMiddleware.preHook()(CTX_DAEMON_EMPTY)
    ]);
  });
  afterEach(function () {
    this.sandbox.restore();
  });
  describe('execute', function () {
    it('calls _execute with inputs', function () {
      this.sandbox.stub(envs, '_prompt').returns(Promise.resolve());
      this.sandbox.stub(envs, '_execute').returns(Promise.resolve());
      return envs.execute(CTX).then(function () {
        sinon.assert.calledOnce(envs._execute);
      });
    });
    it('prompts for missing inputs', function () {
      this.sandbox.stub(envs, '_prompt').returns(Promise.resolve());
      return envs.execute(CTX).then(function () {
        sinon.assert.calledOnce(envs._prompt);
      });
    });
    it('skips the prompt when inputs are supplied', function () {
      this.sandbox.spy(envs, '_prompt');
      this.sandbox.stub(envs, '_execute').returns(Promise.resolve());
      CTX.params = [ENV.body.name];
      CTX.options = { service: { value: SERVICE.body.name } };
      return envs.execute(CTX).then(function () {
        sinon.assert.notCalled(envs._prompt);
      });
    });
    it('prompt - converts service name to owner_id before POST', function () {
      this.sandbox.stub(envs, '_execute').returns(Promise.resolve());
      this.sandbox.stub(envs, '_prompt').returns(Promise.resolve({
        name: ENV.body.name,
        service: SERVICE.body.name
      }));
      return envs.execute(CTX).then(function () {
        sinon.assert.calledOnce(envs._prompt);
        var args = envs._execute.firstCall.args;
        assert.deepEqual(args[1], {
          name: ENV.body.name,
          owner_id: SERVICE.id
        });
      });
    });
    it('rejects invalid flags', function (done) {
      this.sandbox.spy(envs, '_prompt');
      this.sandbox.stub(envs, '_execute').returns(Promise.resolve());
      CTX.params = [ENV.body.name];
      CTX.options = { service: { value: 'a+++' } };
      envs.execute(CTX).then(function () {
        done(new Error('not called'));
      }).catch(function (err) {
        assert.ok(err instanceof ValidationError);
        done();
      });
    });
    it('rejects invalid params', function (done) {
      this.sandbox.spy(envs, '_prompt');
      this.sandbox.stub(envs, '_execute').returns(Promise.resolve());
      CTX.params = ['A+++'];
      CTX.options = { service: { value: SERVICE.body.name } };
      envs.execute(CTX).then(function () {
        done(new Error('not called'));
      }).catch(function (err) {
        assert.ok(err instanceof ValidationError);
        done();
      });
    });
    it('no prompt - converts service name to owner_id before POST', function () {
      this.sandbox.spy(envs, '_prompt');
      this.sandbox.stub(envs, '_execute').returns(Promise.resolve());
      CTX.params = [ENV.body.name];
      CTX.options = { service: { value: SERVICE.body.name } };
      return envs.execute(CTX).then(function () {
        sinon.assert.notCalled(envs._prompt);
        var args = envs._execute.firstCall.args;
        assert.deepEqual(args[1], {
          name: ENV.body.name,
          owner_id: SERVICE.id
        });
      });
    });
  });
  describe('_execute', function () {
    it('authorizes the client', function () {
      var input = { name: 'staging', owner_id: SERVICE.id };
      return envs._execute(CTX.token, input).then(function () {
        sinon.assert.calledOnce(client.auth);
      });
    });
    it('fails if token not found in daemon', function (done) {
      var input = { name: 'staging', owner_id: SERVICE.id };
      envs._execute(CTX_DAEMON_EMPTY.token, input).then(function () {
        done(new Error('dont call'));
      }).catch(function (err) {
        assert.equal(err.message, 'must authenticate first');
        done();
      });
    });
    it('sends api request to envs', function () {
      var input = { name: 'staging', owner_id: SERVICE.id };
      return envs._execute(CTX.token, input).then(function () {
        sinon.assert.calledOnce(client.post);
        var firstCall = client.post.firstCall;
        var args = firstCall.args;
        assert.deepEqual(args[0], {
          url: '/envs',
          json: {
            body: {
              name: 'staging',
              owner_id: SERVICE.id
            }
          }
        });
      });
    });
  });
});
