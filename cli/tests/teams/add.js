/* eslint-env mocha */

'use strict';

var sinon = require('sinon');
var assert = require('assert');
var utils = require('common/utils');
var Promise = require('es6-promise').Promise;

var teamsAdd = require('../../lib/teams/add');
var Session = require('../../lib/session');
var api = require('../../lib/api');
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
    // Context stub with session set
    ctx = new Context({});
    ctx.config = new Config(process.cwd());
    ctx.session = new Session({ token: 'aa', passphrase: 'dd' });
    ctx.api = api.build({ auth_token: ctx.session.token });
    ctx.daemon = new Daemon(ctx.config);
    ctx.params = [PROFILE.body.username, TEAM.body.name];
    ctx.options = {
      org: { value: ORG.body.name }
    };
    ctx.target = new Target({
      path: process.cwd(),
      context: null
    });

    this.sandbox.stub(teamsAdd.output, 'success');
    this.sandbox.stub(teamsAdd.output, 'failure');
    this.sandbox.stub(ctx.api.memberships, 'create')
      .returns(Promise.resolve(MEMBERSHIP));
    this.sandbox.stub(ctx.api.users, 'profile')
      .returns([PROFILE]);
    this.sandbox.stub(ctx.api.orgs, 'get')
      .returns([ORG]);
    this.sandbox.stub(ctx.api.memberships, 'get')
      .returns([MEMBERSHIP]);
    this.sandbox.stub(ctx.api.teams, 'get')
      .returns([TEAM]);
  });

  afterEach(function () {
    this.sandbox.restore();
  });

  describe('#execute', function () {
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
        sinon.assert.calledWith(ctx.api.memberships.create, {
          org_id: ORG.id,
          owner_id: PROFILE.id,
          team_id: TEAM.id
        });
      });
    });

    [
      {
        stub: function () {
          ctx.api.users.profile.returns(Promise.resolve(null));
        },
        error: 'user not found: ' + PROFILE.body.username
      },
      {
        stub: function () {
          ctx.api.orgs.get.returns(Promise.resolve(null));
        },
        error: 'org not found: ' + ORG.body.name
      },
      {
        stub: function () {
          ctx.api.teams.get.returns(Promise.resolve(null));
        },
        error: 'team not found: ' + TEAM.body.name
      }
    ].map(function (testProps) {
      return it('errors: ' + testProps.error, function () {
        testProps.stub();
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
