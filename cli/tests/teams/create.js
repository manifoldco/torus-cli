/* eslint-env mocha */

'use strict';

var sinon = require('sinon');
var assert = require('assert');
var utils = require('common/utils');
var Promise = require('es6-promise').Promise;

var teamsCreate = require('../../lib/teams/create');
var api = require('../../lib/api');
var Config = require('../../lib/config');
var Context = require('../../lib/cli/context');
var Target = require('../../lib/context/target');

var USER_TYPE = 'user';

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

describe('Team Create', function () {
  var ctx;

  before(function () {
    this.sandbox = sinon.sandbox.create();
  });

  beforeEach(function () {
    // Context stub
    ctx = new Context({});
    ctx.config = new Config(process.cwd());
    ctx.api = api.build();
    ctx.params = [TEAM.body.name];
    ctx.options = {
      org: { value: ORG.body.name }
    };
    ctx.target = new Target({
      path: process.cwd(),
      context: null
    });

    this.sandbox.stub(teamsCreate.output, 'success');
    this.sandbox.stub(teamsCreate.output, 'failure');
    this.sandbox.stub(teamsCreate, '_prompt')
      .returns(Promise.resolve({
        org: ORG.body.name
      }));
    this.sandbox.stub(ctx.api.teams, 'create')
      .returns(Promise.resolve(TEAM));
    this.sandbox.stub(ctx.api.orgs, 'get')
      .returns(Promise.resolve([ORG]));
  });

  afterEach(function () {
    this.sandbox.restore();
  });

  describe('#execute', function () {
    it('errors if res from /orgs is invalid', function () {
      ctx.api.orgs.get.returns(Promise.resolve([]));
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
        sinon.assert.calledWith(ctx.api.teams.create, {
          org_id: ORG.id,
          name: TEAM.body.name,
          type: USER_TYPE
        });
      });
    });
  });
});
