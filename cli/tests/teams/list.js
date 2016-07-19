/* eslint-env mocha */

'use strict';

var sinon = require('sinon');
var assert = require('assert');
var utils = require('common/utils');
var Promise = require('es6-promise').Promise;

var teamsList = require('../../lib/teams/list');
var api = require('../../lib/api');
var Config = require('../../lib/config');
var Context = require('../../lib/cli/context');
var Target = require('../../lib/context/target');

var SELF = {
  id: utils.id('user'),
  version: 1,
  body: {
    username: 'skywalker'
  }
};

var ORG = {
  id: utils.id('org'),
  version: 1,
  body: {
    name: 'rebel-alliance'
  }
};

var TEAM = {
  id: utils.id('team'),
  version: 1,
  body: {
    name: 'fighter-squad'
  }
};

var MEMBERSHIP = {
  id: utils.id('membership'),
  version: 1,
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
    this.sandbox.stub(teamsList.output, 'success');
    this.sandbox.stub(teamsList.output, 'failure');

    // Context stub
    ctx = new Context({});
    ctx.config = new Config(process.cwd());
    ctx.api = api.build();
    ctx.params = [];
    ctx.options = {
      org: { value: ORG.body.name }
    };
    ctx.target = new Target({
      path: process.cwd(),
      context: null
    });

    this.sandbox.stub(ctx.api.users, 'self')
      .returns(Promise.resolve([SELF]));
    this.sandbox.stub(ctx.api.orgs, 'get')
      .returns(Promise.resolve([ORG]));
    this.sandbox.stub(ctx.api.memberships, 'get')
      .returns(Promise.resolve([MEMBERSHIP]));
    this.sandbox.stub(ctx.api.teams, 'get')
      .returns(Promise.resolve([TEAM]));
  });

  afterEach(function () {
    this.sandbox.restore();
  });

  describe('#execute', function () {
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
        stub: function () {
          ctx.api.users.self.returns(Promise.resolve([]));
        },
        error: 'current user could not be retrieved'
      },
      {
        stub: function () {
          ctx.api.orgs.get.returns(Promise.resolve([]));
        },
        error: 'org not found: ' + ORG.body.name
      },
      {
        stub: function () {
          ctx.api.teams.get.returns(Promise.resolve(null));
        },
        error: 'could not find team(s)'
      },
      {
        stub: function () {
          ctx.api.memberships.get.returns(Promise.resolve(null));
        },
        error: 'could not find memberships'
      }
    ].map(function (testProps) {
      return it('errors: ' + testProps.error, function () {
        testProps.stub();
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
