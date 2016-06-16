/* eslint-env mocha */

'use strict';

var sinon = require('sinon');
var assert = require('assert');
var utils = require('common/utils');
var Promise = require('es6-promise').Promise;

var serviceCreate = require('../../lib/services/create');
var client = require('../../lib/api/client').create();
var Config = require('../../lib/config');
var Context = require('../../lib/cli/context');
var Target = require('../../lib/context/target');
var Daemon = require('../../lib/daemon/object').Daemon;
var sessionMiddleware = require('../../lib/middleware/session');

var ORG = {
  id: utils.id('org'),
  body: {
    name: 'my-org'
  }
};

var PROJECT = {
  id: utils.id('project'),
  body: {
    name: 'api',
    org_id: ORG.id
  }
};

var SERVICE = {
  id: utils.id('service'),
  body: {
    name: 'www',
    project_id: PROJECT.id,
    org_id: ORG.id
  }
};

describe('Services Create', function () {
  var ctx;

  before(function () {
    this.sandbox = sinon.sandbox.create();
  });

  beforeEach(function () {
    this.sandbox.stub(serviceCreate.output, 'success');
    this.sandbox.stub(serviceCreate.output, 'failure');
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
      .returns(Promise.resolve({
        body: [SERVICE]
      }));
    this.sandbox.spy(client, 'auth');

    // Context stub with token set
    ctx = new Context({});
    ctx.config = new Config(process.cwd());
    ctx.daemon = new Daemon(ctx.config);
    ctx.params = ['abc123abc'];
    ctx.options = {
      org: { value: ORG.body.name },
      project: { value: PROJECT.body.name }
    };
    ctx.target = new Target(process.cwd(), {});

    // Daemon with token
    this.sandbox.stub(ctx.daemon, 'set')
      .returns(Promise.resolve());
    this.sandbox.stub(ctx.daemon, 'get')
      .returns(Promise.resolve({
        token: 'this is a token',
        passphrase: 'a passphrase'
      }));

    // Run the token middleware to populate the context object
    return sessionMiddleware()(ctx);
  });

  afterEach(function () {
    this.sandbox.restore();
  });

  describe('execute', function () {
    it('errors if org is not provided', function () {
      ctx.options.org.value = undefined;

      return serviceCreate.execute(ctx).then(function () {
        assert.ok(false, 'should not pass');
      }, function (err) {
        assert.ok(err);
        assert.strictEqual(err.message, '--org is required.');
      });
    });

    it('errors if org does not exist', function () {
      client.get.onCall(0).returns(Promise.resolve({ body: [] }));

      return serviceCreate.execute(ctx).then(function () {
        assert.ok(false, 'should not pass');
      }, function (err) {
        assert.ok(err);
        assert.strictEqual(err.message, 'org not found: my-org');
      });
    });

    it('errors if project specified and not found', function () {
      ctx.options.project.value = 'api';
      client.get.onCall(1).returns(Promise.resolve({ body: [] }));

      return serviceCreate.execute(ctx).then(function () {
        assert.ok(false, 'should not pass');
      }, function (err) {
        assert.ok(err);
        assert.strictEqual(err.message, 'project not found: api');
      });
    });

    it('errors if project doesnt exist', function () {
      ctx.options.project.value = undefined;
      client.get.onCall(1).returns(Promise.resolve({ body: [] }));

      return serviceCreate.execute(ctx).then(function () {
        assert.ok(false, 'should error');
      }, function (err) {
        assert.ok(err);
        assert.strictEqual(err.message,
          'You must create a project before creating a service');
      });
    });

    it('errors if properties provided and are invalid', function () {
      ctx.options.project.value = '--df!';
      ctx.options.org.value = 'my-org'; // org must be valid
      ctx.params = ['@@2'];

      return serviceCreate.execute(ctx).then(function () {
        assert.ok(false, 'should error');
      }, function (err) {
        assert.ok(err);
        assert.strictEqual(err.message,
          'name: Only alphanumeric, hyphens and underscores are allowed');
      });
    });

    it('creates service if options provided', function () {
      this.sandbox.stub(serviceCreate, '_execute').returns(Promise.resolve({
        project: PROJECT,
        service: SERVICE
      }));

      return serviceCreate.execute(ctx).then(function (result) {
        assert.deepEqual(result, {
          project: PROJECT,
          service: SERVICE
        });

        sinon.assert.calledWith(serviceCreate._execute, ORG, [PROJECT], {
          name: ctx.params[0],
          project: PROJECT.body.name,
          org: ORG.body.name
        });
      });
    });

    it('creates srevice if prompted', function () {
      ctx.params = [];
      ctx.options.project.value = undefined;

      this.sandbox.stub(serviceCreate, '_prompt').returns(Promise.resolve({
        name: SERVICE.body.name,
        project: PROJECT.body.name
      }));

      return serviceCreate.execute(ctx).then(function (result) {
        assert.deepEqual(result, {
          service: SERVICE,
          project: PROJECT
        });
      });
    });
  });

  describe('_execute', function () {
    var data = {
      name: SERVICE.body.name,
      project: PROJECT.body.name
    };

    it('errors if project not found', function () {
      return serviceCreate._execute(ORG, [], data).then(function () {
        assert.ok(false, 'should error');
      }, function (err) {
        assert.ok(err);
        assert.strictEqual(err.message, 'project not found: api');
      });
    });

    it('returns error if api returns error', function () {
      client.post.onCall(0).returns(Promise.reject(new Error('bad')));

      return serviceCreate._execute(ORG, [PROJECT], data).then(function () {
        assert.ok(false, 'should error');
      }, function (err) {
        assert.ok(err);
        assert.strictEqual(err.message, 'bad');
      });
    });

    it('makes service object', function () {
      return serviceCreate._execute(ORG, [PROJECT], data)
      .then(function (result) {
        assert.deepEqual(result, {
          project: PROJECT,
          service: SERVICE
        });
      });
    });
  });
});
