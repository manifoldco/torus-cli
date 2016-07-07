/* eslint-env mocha */

'use strict';

var sinon = require('sinon');
var assert = require('assert');
var utils = require('common/utils');
var Promise = require('es6-promise').Promise;

var teamsCreate = require('../../lib/teams/create');
var client = require('../../lib/api/client').create();
var sessionMiddleware = require('../../lib/middleware/session');
var Config = require('../../lib/config');
var Context = require('../../lib/cli/context');
var Target = require('../../lib/context/target');
var Daemon = require('../../lib/daemon/object').Daemon;

var USER_TYPE = 'user';

var ORG = {
  id: utils.id('org'),
  body: {
    name: 'rebel-alliance'
  }
};

var TEAM = {
  id: utils.id('team'),
  body: {
    name: 'fighter-squad'
  }
};

describe('Team Create', function () {
  var ctx;

  before(function () {
    this.sandbox = sinon.sandbox.create();
  });

  beforeEach(function () {
    this.sandbox.spy(client, 'auth');
    this.sandbox.stub(teamsCreate.output, 'success');
    this.sandbox.stub(teamsCreate.output, 'failure');
    this.sandbox.stub(teamsCreate, '_prompt')
      .returns(Promise.resolve({
        org: ORG.body.name
      }));
    this.sandbox.stub(client, 'post')
      .returns(Promise.resolve({ body: TEAM }));
    this.sandbox.stub(client, 'get')
      .withArgs({
        url: '/orgs'
      })
      .returns(Promise.resolve({
        body: [ORG]
      }));

    // Context stub with session set
    ctx = new Context({});
    ctx.config = new Config(process.cwd());
    ctx.daemon = new Daemon(ctx.config);
    ctx.params = [TEAM.body.name];
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
    it('verify\'s the session', function () {
      return teamsCreate.execute(ctx).then(function () {
        sinon.assert.calledOnce(client.auth);
      });
    });

    it('errors if res from /orgs is invalid', function () {
      client.get.withArgs({
        url: '/orgs'
      })
      .returns(Promise.resolve({
        body: []
      }));

      return teamsCreate.execute(ctx).then(function () {
        assert.ok(false, 'should error');
      }, function (err) {
        assert.ok(err);
        assert.strictEqual(err.message, 'Could not locate organizations');
      });
    });

    it('errors if team [name] is invalid', function () {
      ctx.params = ['rebel @lliance'];

      return teamsCreate.execute(ctx).then(function () {
        assert.ok(false, 'should error');
      }, function (err) {
        var errMessage = 'name: Only alphanumeric, hyphens and underscores are allowed';
        assert.ok(err);
        assert.strictEqual(err.message, errMessage);
      });
    });

    it('errors if org [name] is invalid', function () {
      ctx.options = { org: { value: 'D@rth Vader' } };

      return teamsCreate.execute(ctx).then(function () {
        assert.ok(false, 'should error');
      }, function (err) {
        var errMessage = 'org: Only alphanumeric, hyphens and underscores are allowed';
        assert.ok(err);
        assert.strictEqual(err.message, errMessage);
      });
    });

    it('errors if org cannot be found', function () {
      ctx.options = { org: { value: 'lost-and-found' } };

      return teamsCreate.execute(ctx).then(function () {
        assert.ok(false, 'should error');
      }, function (err) {
        var errMessage = 'unknown org: lost-and-found';
        assert.ok(err);
        assert.strictEqual(err.message, errMessage);
      });
    });

    it('creates a new team object', function () {
      return teamsCreate.execute(ctx).then(function (team) {
        assert.deepEqual(team, TEAM);
        sinon.assert.calledWith(client.post, {
          url: '/teams',
          json: {
            body: {
              org_id: ORG.id,
              name: TEAM.body.name,
              type: USER_TYPE
            }
          }
        });
      });
    });
  });
});
