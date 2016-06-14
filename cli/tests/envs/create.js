/* eslint-env mocha */
'use strict';

var sinon = require('sinon');
var assert = require('assert');
var utils = require('common/utils');
var Promise = require('es6-promise').Promise;

var listServices = require('../../lib/services/list');
var envs = require('../../lib/envs/create');
var client = require('../../lib/api/client').create();
var sessionMiddleware = require('../../lib/middleware/session');
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

var PROJECT = {
  id: utils.id('project'),
  body: {
    name: 'api-1',
    org_id: ORG.id
  }
};

var SERVICE = {
  id: utils.id('service'),
  body: {
    name: 'api-1',
    project_id: PROJECT.id,
    org_id: ORG.id
  }
};

var ENV = {
  id: utils.id('env'),
  body: {
    name: 'staging',
    project_id: PROJECT.id,
    org_id: ORG.id
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
      .onFirstCall()
      .returns(Promise.resolve({
        body: [ORG]
      }))
      .onSecondCall()
      .returns(Promise.resolve({
        body: [SERVICE]
      }));
    this.sandbox.stub(client, 'post')
      .returns(Promise.resolve({ body: [ENV] }));
    this.sandbox.spy(client, 'auth');
    this.sandbox.stub(listServices, 'execute')
      .returns(Promise.resolve({ body: [SERVICE] }));

    // Context stub when no session set
    CTX_DAEMON_EMPTY = new Context({});
    CTX_DAEMON_EMPTY.config = new Config(process.cwd());
    CTX_DAEMON_EMPTY.daemon = new Daemon(CTX_DAEMON_EMPTY.config);

    // Context stub with session set
    CTX = new Context({});
    CTX.config = new Config(process.cwd());
    CTX.daemon = new Daemon(CTX.config);
    CTX.params = ['abc123abc'];
    CTX.options = {
      service: { value: SERVICE.body.name },
      org: { value: ORG.body.name }
    };

    // Empty daemon
    this.sandbox.stub(CTX_DAEMON_EMPTY.daemon, 'set')
      .returns(Promise.resolve());
    this.sandbox.stub(CTX_DAEMON_EMPTY.daemon, 'get')
      .returns(Promise.resolve({ token: '', passphrase: '' }));

    // Daemon with token
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
  describe('execute', function () {
    it('calls _execute with inputs', function () {
      this.sandbox.stub(envs, '_execute').returns(Promise.resolve());
      return envs.execute(CTX).then(function () {
        sinon.assert.calledOnce(envs._execute);
      });
    });
    it('errors if given service not found', function () {
      CTX.params = ['dev'];
      CTX.options.service = { value: 'api-2' };
      return envs.execute(CTX).then(function () {
        assert.ok(false, 'should error');
      }).catch(function (err) {
        assert.ok(err.message, 'Unknown service: api-2');
      });
    });
    it('skips the prompt when inputs are supplied', function () {
      this.sandbox.spy(envs, '_prompt');
      this.sandbox.stub(envs, '_execute').returns(Promise.resolve());
      CTX.params = [ENV.body.name];
      CTX.options = {
        service: { value: SERVICE.body.name },
        org: { value: ORG.body.name }
      };
      return envs.execute(CTX).then(function () {
        sinon.assert.notCalled(envs._prompt);
      });
    });
    it('rejects invalid flags', function (done) {
      this.sandbox.spy(envs, '_prompt');
      this.sandbox.stub(envs, '_execute').returns(Promise.resolve());
      CTX.params = [ENV.body.name];
      CTX.options = {
        service: { value: 'a+++' },
        org: { value: ORG.body.name }
      };
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
      CTX.options = {
        service: { value: SERVICE.body.name },
        org: { value: ORG.body.name }
      };
      envs.execute(CTX).then(function () {
        done(new Error('not called'));
      }).catch(function (err) {
        assert.ok(err instanceof ValidationError);
        done();
      });
    });
    it('no prompt - converts service name to project_id before POST', function () {
      this.sandbox.spy(envs, '_prompt');
      this.sandbox.stub(envs, '_execute').returns(Promise.resolve());
      CTX.params = [ENV.body.name];
      CTX.options = {
        service: { value: SERVICE.body.name },
        org: { value: ORG.body.name }
      };
      return envs.execute(CTX).then(function () {
        sinon.assert.notCalled(envs._prompt);
        var args = envs._execute.firstCall.args;
        assert.deepEqual(args[1], {
          body: {
            name: ENV.body.name,
            project_id: PROJECT.id,
            org_id: ORG.id
          }
        });
      });
    });
  });
  describe('_execute', function () {
    it('fails if session does not exist', function (done) {
      var input = {
        body: {
          name: 'staging',
          project_id: PROJECT.id
        }
      };
      envs._execute(CTX_DAEMON_EMPTY.session, input).then(function () {
        done(new Error('dont call'));
      }).catch(function (err) {
        assert.equal(err.message, 'Session object missing on Context');
        done();
      });
    });

    it('sends api request to envs', function () {
      var input = {
        body: {
          name: 'staging',
          project_id: PROJECT.id
        }
      };
      return envs._execute(CTX.session, input).then(function () {
        sinon.assert.calledOnce(client.post);
        var firstCall = client.post.firstCall;
        var args = firstCall.args;
        assert.deepEqual(args[0], {
          url: '/envs',
          json: {
            body: {
              name: 'staging',
              project_id: PROJECT.id
            }
          }
        });
      });
    });
  });
});
