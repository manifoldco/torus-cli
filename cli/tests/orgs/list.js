/* eslint-env mocha */

'use strict';

var sinon = require('sinon');
var assert = require('assert');
var utils = require('common/utils');
var Promise = require('es6-promise').Promise;

var orgsList = require('../../lib/orgs/list');
var client = require('../../lib/api/client').create();
var sessionMiddleware = require('../../lib/middleware/session');
var Config = require('../../lib/config');
var Context = require('../../lib/cli/context');
var Target = require('../../lib/context/target');
var Daemon = require('../../lib/daemon/object').Daemon;

var ORG = {
  id: utils.id('org'),
  body: {
    name: 'jeff-arigato-sh'
  }
};

var SELF = {
  id: utils.id('user'),
  body: {
    username: 'skywalker'
  }
};

describe('Orgs List', function () {
  var ctx;

  before(function () {
    this.sandbox = sinon.sandbox.create();
  });

  beforeEach(function () {
    this.sandbox.stub(orgsList.output, 'success');
    this.sandbox.stub(orgsList.output, 'failure');
    this.sandbox.stub(client, 'get')
      .withArgs({ url: '/users/self' })
      .returns(Promise.resolve({
        body: [SELF]
      }))
      .withArgs({ url: '/orgs' })
      .returns(Promise.resolve({
        body: [ORG]
      }));
    this.sandbox.spy(client, 'auth');

    // Context stub with session set
    ctx = new Context({});
    ctx.config = new Config(process.cwd());
    ctx.daemon = new Daemon(ctx.config);
    ctx.params = [];
    ctx.options = {
      org: { value: ORG.body.name }
    };
    ctx.target = new Target({
      path: process.cwd(),
      context: null
    });

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
    it('returns orgs', function () {
      return orgsList.execute(ctx).then(function (results) {
        assert.deepEqual(results, {
          self: SELF,
          orgs: [ORG]
        });
      });
    });

    it('errors if session is missing on ctx', function () {
      ctx.session = null;

      return orgsList.execute(ctx).then(function () {
        assert.ok(false, 'should error');
      }, function (err) {
        assert.ok(err);
        assert.strictEqual(err.message, 'Session object missing on Context');
      });
    });

    it('errors if the org was not found', function () {
      client.get
        .withArgs({ url: '/users/self' })
        .returns(Promise.resolve({
          body: []
        }))
        .withArgs({ url: '/orgs' })
        .returns(Promise.resolve({
          body: []
        }));

      return orgsList.execute(ctx).then(function () {
        assert.ok(false, 'should error');
      }, function (err) {
        assert.ok(err);
        assert.strictEqual(err.message, 'No orgs found');
      });
    });
  });
});
