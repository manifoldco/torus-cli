/* eslint-env mocha */
'use strict';

var sinon = require('sinon');
var assert = require('assert');
var utils = require('common/utils');
var Promise = require('es6-promise').Promise;

var envCreate = require('../../lib/envs/create');
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

var ENV = {
  id: utils.id('env'),
  body: {
    name: 'staging',
    project_id: PROJECT.id,
    org_id: ORG.id
  }
};

var ctx_DAEMON_EMPTY;
var ctx;

describe('Envs Create', function () {
  var ctx;

  before(function () {
    this.sandbox = sinon.sandbox.create();
  });

  beforeEach(function () {
    this.sandbox.stub(envCreate.output, 'success');
    this.sandbox.stub(envCreate.output, 'failure');
    this.sandbox.stub(client, 'get')
      .onFirstCall()
      .returns(Promise.resolve({
        body: [ORG]
      }))
      .onSecondCall()
      .returns(Promise.resolve({
        body: [PROJECT]
      }));
    
    this.sandbox.stub(client, 'post')
      .returns(Promise.resolve({ body: [ENV] }));
    this.sandbox.spy(client, 'auth');
    
    // Context stub with session set
    ctx = new Context({});
    ctx.config = new Config(process.cwd());
    ctx.daemon = new Daemon(ctx.config);
    ctx.params = ['abc123abc'];
    ctx.options = {
      project: { value: PROJECT.body.name },
      org: { value: ORG.body.name }
    };

    // Daemon with token
    this.sandbox.stub(ctx.daemon, 'set')
      .returns(Promise.resolve());
    this.sandbox.stub(ctx.daemon, 'get')
      .returns(Promise.resolve({ token: 'this is a token', passphrase: 'hi' }));

    return sessionMiddleware()(ctx)  
  });
  afterEach(function () {
    this.sandbox.restore();
  });
  describe('execute', function () {
    
    it('errors if org is not provided', function () {
      ctx.options.org.value = undefined;

      return envCreate.execute(ctx).then(function () {
        assert.ok(false, 'should not pass');
      }, function (err) {
        assert.ok(err);
        assert.strictEqual(err.message, '--org is required.');
      });
    });

    it('errors if org does not exist', function () {
      client.get.onCall(0).returns(Promise.resolve({ body: [] }));

      return envCreate.execute(ctx).then(function () {
        assert.ok(false, 'should not pass');
      }, function (err) {
        assert.ok(err);
        assert.strictEqual(err.message, 'org not found: jeff-arigato-sh');
      });
    });

    it('errors if project specified and not found', function () {
      ctx.options.project.value = 'api';
      client.get.onCall(1).returns(Promise.resolve({ body: [] }));

      return envCreate.execute(ctx).then(function () {
        assert.ok(false, 'should not pass');
      }, function (err) {
        assert.ok(err);
        assert.strictEqual(err.message, 'project not found: api');
      });
    });

    it('errors if project doesnt exist', function () {
      ctx.options.project.value = undefined;
      client.get.onCall(1).returns(Promise.resolve({ body: [] }));

      return envCreate.execute(ctx).then(function () {
        assert.ok(false, 'should error');
      }, function (err) {
        assert.ok(err);
        assert.strictEqual(err.message,
          'You must create a project before creating an environment');
      });
    });

    it('errors if properties provided and are invalid', function () {
      ctx.options.project.value = '--df!';
      ctx.options.org.value = 'my-org'; // org must be valid
      ctx.params = ['@@2'];

      return envCreate.execute(ctx).then(function () {
        assert.ok(false, 'should error');
      }, function (err) {
        assert.ok(err);
        assert.strictEqual(err.message,
          'name: Only alphanumeric, hyphens and underscores are allowed');
      });
    });

    it('creates env if options provided', function () {
      this.sandbox.stub(envCreate, '_execute').returns(Promise.resolve({
        project: PROJECT,
        env: ENV
      }));

      return envCreate.execute(ctx).then(function (result) {
        assert.deepEqual(result, {
          project: PROJECT,
          env: ENV
        });

        sinon.assert.calledWith(envCreate._execute, ORG, [PROJECT], {
          name: ctx.params[0],
          project: PROJECT.body.name,
          org: ORG.body.name
        });
      });
    });

    it('creates env if prompted', function () {
      ctx.params = [];
      ctx.options.project.value = undefined;

      this.sandbox.stub(envCreate, '_prompt').returns(Promise.resolve({
        name: ENV.body.name,
        project: PROJECT.body.name
      }));

      return envCreate.execute(ctx).then(function (result) {
        assert.deepEqual(result, {
          env: ENV,
          project: PROJECT
        });
      });
    });
  });

  describe('_execute', function () {
    var data = {
      name: ENV.body.name,
      project: PROJECT.body.name
    };

    it('errors if project not found', function () {
      return envCreate._execute(ORG, [], data).then(function () {
        assert.ok(false, 'should error');
      }, function (err) {
        assert.ok(err);
        assert.strictEqual(err.message, 'project not found: api-1');
      });
    });

    it('returns error if api returns error', function () {
      client.post.onCall(0).returns(Promise.reject(new Error('bad')));

      return envCreate._execute(ORG, [PROJECT], data).then(function () {
        assert.ok(false, 'should error');
      }, function (err) {
        assert.ok(err);
        assert.strictEqual(err.message, 'bad');
      });
    });

    it('makes service object', function () {
      return envCreate._execute(ORG, [PROJECT], data)
      .then(function (result) {
        assert.deepEqual(result, {
          project: PROJECT,
          env: ENV
        });
      });
    });
  });
});
