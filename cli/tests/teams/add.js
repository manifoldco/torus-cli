/* eslint-env mocha */

'use strict';

var sinon = require('sinon');
var assert = require('assert');
var utils = require('common/utils');
var Promise = require('es6-promise').Promise;

var teamsAdd = require('../../lib/teams/add');
var client = require('../../lib/api/client').create();
var sessionMiddleware = require('../../lib/middleware/session');
var Config = require('../../lib/config');
var Context = require('../../lib/cli/context');
var Target = require('../../lib/context/target');
var Daemon = require('../../lib/daemon/object').Daemon;

var PROFILE = {
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
    owner_id: PROFILE.id,
    org_id: ORG.id,
    team_id: TEAM.id
  }
};

describe('Team Add', function () {
  var ctx;

  before(function () {
    this.sandbox = sinon.sandbox.create();
  });

  beforeEach(function () {
    this.sandbox.spy(client, 'auth');
    this.sandbox.stub(teamsAdd.output, 'success');
    this.sandbox.stub(teamsAdd.output, 'failure');
    this.sandbox.stub(client, 'post')
      .returns(Promise.resolve({ body: MEMBERSHIP }));
    this.sandbox.stub(client, 'get')
      .withArgs({ url: '/profiles/' + PROFILE.body.username })
      .returns(Promise.resolve({
        body: [PROFILE]
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
          owner_id: PROFILE.id
        }
      })
      .returns(Promise.resolve({
        body: [MEMBERSHIP]
      }))
      .withArgs({
        url: '/teams',
        qs: {
          org_id: ORG.id,
          name: TEAM.body.name
        }
      })
      .returns(Promise.resolve({
        body: [TEAM]
      }));

    // Context stub with session set
    ctx = new Context({});
    ctx.config = new Config(process.cwd());
    ctx.daemon = new Daemon(ctx.config);
    ctx.params = [PROFILE.body.username, TEAM.body.name];
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
      return teamsAdd.execute(ctx).then(function () {
        sinon.assert.calledOnce(client.auth);
      });
    });

    it('errors without context and --org [name]', function () {
      ctx.options = { org: { value: undefined } };

      return teamsAdd.execute(ctx).then(function () {
        assert.ok(false, 'should error');
      }, function (err) {
        assert.ok(err);
        assert.strictEqual(err.message, '--org is required.');
      });
    });

    it('errors without [username] and/or [team]', function () {
      ctx.params = [];

      return teamsAdd.execute(ctx).then(function () {
        assert.ok(false, 'should error');
      }, function (err) {
        assert.ok(err);
        assert.strictEqual(err.message, 'username and team are required.');
      });
    });

    it('errors if --org [name] is invalid', function () {
      ctx.options = { org: { value: 'D@rth Vader' } };

      return teamsAdd.execute(ctx).then(function () {
        assert.ok(false, 'should error');
      }, function (err) {
        var errMessage = 'org: Only alphanumeric, hyphens and underscores are allowed';
        assert.ok(err);
        assert.strictEqual(err.message, errMessage);
      });
    });

    it('errors if [username] is invalid', function () {
      ctx.params = ['D@rth Vader', TEAM.body.name];

      return teamsAdd.execute(ctx).then(function () {
        assert.ok(false, 'should error');
      }, function (err) {
        var errMessage = 'username: Only alphanumeric, hyphens and underscores are allowed';
        assert.ok(err);
        assert.strictEqual(err.message, errMessage);
      });
    });

    it('errors if [name] is invalid', function () {
      ctx.params = [PROFILE.body.username, 'D@rth Vader'];

      return teamsAdd.execute(ctx).then(function () {
        assert.ok(false, 'should error');
      }, function (err) {
        var errMessage = 'team: Only alphanumeric, hyphens and underscores are allowed';
        assert.ok(err);
        assert.strictEqual(err.message, errMessage);
      });
    });

    it('requests a profile, orgs, teams', function () {
      return teamsAdd.execute(ctx).then(function (payload) {
        assert.deepEqual(payload, {
          org: ORG,
          team: TEAM,
          user: PROFILE
        });
      });
    });

    it('creates a new membership object', function () {
      return teamsAdd.execute(ctx).then(function () {
        sinon.assert.calledWith(client.post, {
          url: '/memberships',
          json: {
            body: {
              org_id: ORG.id,
              owner_id: PROFILE.id,
              team_id: TEAM.id
            }
          }
        });
      });
    });

    [
      {
        url: '/profiles/' + PROFILE.body.username,
        error: 'user not found: ' + PROFILE.body.username
      },
      {
        url: '/orgs',
        error: 'org not found: ' + ORG.body.name,
        qs: { name: ORG.body.name }
      },
      {
        url: '/teams',
        error: 'team not found: ' + TEAM.body.name,
        qs: {
          org_id: ORG.id,
          name: TEAM.body.name
        }
      }
    ].map(function (testProps) {
      var req = { url: testProps.url };

      if (testProps.qs) {
        req.qs = testProps.qs;
      }

      return it('errors if res from ' + req.url + ' is invalid', function () {
        client.get.withArgs(req).returns(Promise.resolve({ body: null }));

        return teamsAdd.execute(ctx).then(function () {
          assert.ok(false, 'should error');
        }, function (err) {
          assert.ok(err);
          assert.strictEqual(err.message, testProps.error);
        });
      });
    });
  });
});
