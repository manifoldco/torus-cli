/* eslint-env mocha */

'use strict';

var sinon = require('sinon');
var assert = require('assert');
var utils = require('common/utils');
var Promise = require('es6-promise').Promise;

var envsList = require('../../lib/envs/list');
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

var PROJECTS = [
  {
    id: utils.id('project'),
    body: {
      name: 'www',
      org_id: ORG.id
    }
  },
  {
    id: utils.id('project'),
    body: {
      name: 'core',
      org_id: ORG.id
    }
  }
];

var ENVS = [
  {
    id: utils.id('env'),
    body: {
      org_id: ORG.id,
      project_id: PROJECTS[0].id,
      name: 'dev-1'
    }
  },
  {
    id: utils.id('env'),
    body: {
      org_id: ORG.id,
      project_id: PROJECTS[1].id,
      name: 'dev-2'
    }
  }
];

describe('Envs List', function () {
  var ctx;

  before(function () {
    this.sandbox = sinon.sandbox.create();
  });

  beforeEach(function () {
    this.sandbox.stub(envsList.output, 'success');
    this.sandbox.stub(envsList.output, 'failure');
    this.sandbox.stub(client, 'get')
      .onFirstCall()
      .returns(Promise.resolve({
        body: [ORG]
      }))
      .onSecondCall()
      .returns(Promise.resolve({
        body: PROJECTS
      }))
      .onThirdCall()
      .returns(Promise.resolve({
        body: ENVS
      }));
    this.sandbox.spy(client, 'auth');

    // Context stub with session set
    ctx = new Context({});
    ctx.config = new Config(process.cwd());
    ctx.daemon = new Daemon(ctx.config);
    ctx.params = [];
    ctx.options = {
      org: { value: ORG.body.name },
      project: { value: PROJECTS[0].body.name }
    };

    // Daemon with session
    this.sandbox.stub(ctx.daemon, 'set')
      .returns(Promise.resolve());
    this.sandbox.stub(ctx.daemon, 'get')
      .returns(Promise.resolve({ token: 'this is a token', passphrase: 'hi' }));

    // Run the session middleware to populate the context object
    return sessionMiddleware()(ctx);
  });

  afterEach(function () {
    this.sandbox.restore();
  });

  describe('#execute', function () {
    it('errors if session is missing on ctx', function () {
      ctx.session = null;

      return envsList.execute(ctx).then(function () {
        assert.ok(false, 'should error');
      }, function (err) {
        assert.ok(err);
        assert.strictEqual(err.message, 'Session object missing on Context');
      });
    });

    it('errors if org is not provided', function () {
      ctx.options.org.value = undefined;

      return envsList.execute(ctx).then(function () {
        assert.ok(false, 'should error');
      }, function (err) {
        assert.ok(err);
        assert.strictEqual(err.message, '--org is required.');
      });
    });

    it('errors if org and project are provided and invalid', function () {
      ctx.options.org.value = '@@';
      ctx.options.project.value = '!!2@@';

      return envsList.execute(ctx).then(function () {
        assert.ok(false, 'should error');
      }, function (err) {
        assert.ok(err);
        assert.strictEqual(err.message,
          'org: Only alphanumeric, hyphens and underscores are allowed');
      });
    });

    it('errors if org or project are provided and invalid', function () {
      ctx.options.org.value = 'jeff-arigato-sh';
      ctx.options.project.value = '@@@';

      return envsList.execute(ctx).then(function () {
        assert.ok(false, 'should error');
      }, function (err) {
        assert.ok(err);
        assert.strictEqual(err.message,
          'project: Only alphanumeric, hyphens and underscores are allowed');
      });
    });

    it('errors if the org was not found', function () {
      client.get.onCall(0).returns(Promise.resolve({ body: [] }));

      return envsList.execute(ctx).then(function () {
        assert.ok(false, 'should error');
      }, function (err) {
        assert.ok(err);
        assert.strictEqual(err.message, 'org not found: jeff-arigato-sh');
      });
    });

    it('errors if project provided and not found', function () {
      client.get.onCall(1).returns(Promise.resolve({ body: [] }));

      return envsList.execute(ctx).then(function () {
        assert.ok(false, 'should error');
      }, function (err) {
        assert.ok(err);
        assert.strictEqual(err.message, 'project not found: www');
      });
    });

    it('returns project and its envs if project provided', function () {
      return envsList.execute(ctx).then(function (results) {
        assert.deepEqual(results, {
          projects: [PROJECTS[0]],
          envs: [ENVS[0]]
        });
      });
    });

    it('returns all projects and envs if project is not provided', function () {
      ctx.options.project.value = undefined;

      return envsList.execute(ctx).then(function (results) {
        assert.deepEqual(results, {
          projects: PROJECTS,
          envs: ENVS
        });
      });
    });
  });
});
