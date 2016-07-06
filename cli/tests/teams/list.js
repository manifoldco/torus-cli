/* eslint-env mocha */

'use strict';

var sinon = require('sinon');
var assert = require('assert');
var utils = require('common/utils');
var Promise = require('es6-promise').Promise;

var teamsList = require('../../lib/teams/list');
var client = require('../../lib/api/client').create();
var sessionMiddleware = require('../../lib/middleware/session');
var Config = require('../../lib/config');
var Context = require('../../lib/cli/context');
var Target = require('../../lib/context/target');
var Daemon = require('../../lib/daemon/object').Daemon;

var SELF = {
  id: utils.id('user'),
  body: {
    username: 'skywalker'
  }
};

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

var MEMBERSHIP = {
  id: utils.id('membership'),
  body: {
    owner_id: SELF.id,
    org_id: ORG.id,
    team_id: TEAM.id
  }
};

describe('Team List', function () {
  var ctx;

  before(function () {
    this.sandbox = sinon.sandbox.create();
  });

  beforeEach(function () {
    this.sandbox.spy(client, 'auth');
    this.sandbox.stub(teamsList.output, 'success');
    this.sandbox.stub(teamsList.output, 'failure');
    this.sandbox.stub(client, 'get')
      .withArgs({ url: '/users/self' })
      .returns(Promise.resolve({
        body: [SELF]
      }))
      .withArgs({
        url: '/orgs',
        qs: { name: ORG.body.name }
      })
      .returns(Promise.resolve({
        body: [ORG]
      }))
      .withArgs({
        url: '/memberships',
        qs: {
          org_id: ORG.id,
          owner_id: SELF.id
        }
      })
      .returns(Promise.resolve({
        body: [MEMBERSHIP]
      }))
      .withArgs({
        url: '/teams',
        qs: {
          org_id: ORG.id
        }
      })
      .returns(Promise.resolve({
        body: [TEAM]
      }));

    // Context stub with session set
    ctx = new Context({});
    ctx.config = new Config(process.cwd());
    ctx.daemon = new Daemon(ctx.config);
    ctx.params = [];
    ctx.options = {
      org: { value: ORG.body.name }
    };
    ctx.target = new Target(process.cwd(), {});

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
      return teamsList.execute(ctx).then(function () {
        sinon.assert.calledOnce(client.auth);
      });
    });

    it('errors without context and --org [name]', function () {
      ctx.options = { org: { value: undefined } };

      return teamsList.execute(ctx).then(function () {
        assert.ok(false, 'should error');
      }, function (err) {
        assert.ok(err);
        assert.strictEqual(err.message, '--org is required.');
      });
    });

    it('errors if [name] is invalid', function () {
      ctx.options = { org: { value: 'D@rth Vader' } };

      return teamsList.execute(ctx).then(function () {
        assert.ok(false, 'should error');
      }, function (err) {
        var errMessage = 'org: Only alphanumeric, hyphens and underscores are allowed';
        assert.ok(err);
        assert.strictEqual(err.message, errMessage);
      });
    });

    it('requests teams, memberships, orgs and the current user', function () {
      return teamsList.execute(ctx).then(function (payload) {
        assert.deepEqual(payload, {
          org: ORG,
          self: SELF,
          teams: [TEAM],
          memberships: [MEMBERSHIP]
        });
      });
    });

    [
      {
        url: '/users/self',
        error: 'current user could not be retrieved'
      },
      {
        url: '/orgs',
        error: 'org not found: ' + ORG.body.name,
        qs: { name: ORG.body.name }
      },
      {
        url: '/teams',
        error: 'could not find team(s)',
        qs: { org_id: ORG.id }
      },
      {
        url: '/memberships',
        error: 'could not find memberships',
        qs: {
          org_id: ORG.id,
          owner_id: SELF.id
        }
      }
    ].map(function (testProps) {
      var req = { url: testProps.url };

      if (testProps.qs) {
        req.qs = testProps.qs;
      }

      return it('errors if res from ' + req.url + ' is invalid', function () {
        client.get.withArgs(req).returns(Promise.resolve({ body: null }));

        return teamsList.execute(ctx).then(function () {
          assert.ok(false, 'should error');
        }, function (err) {
          assert.ok(err);
          assert.strictEqual(err.message, testProps.error);
        });
      });
    });
  });
});
