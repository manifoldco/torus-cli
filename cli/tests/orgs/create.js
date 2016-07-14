/* eslint-env mocha */
'use strict';

var sinon = require('sinon');
var assert = require('assert');
var utils = require('common/utils');
var Promise = require('es6-promise').Promise;

var orgCreate = require('../../lib/orgs/create');
var Config = require('../../lib/config');
var Context = require('../../lib/cli/context');
var Target = require('../../lib/context/target');
var Daemon = require('../../lib/daemon/object').Daemon;
var Session = require('../../lib/session');
var api = require('../../lib/api');

var ORG = {
  id: utils.id('org'),
  version: 1,
  body: {
    name: 'knotty-buoy'
  }
};

describe('Orgs Create', function () {
  var ctx;

  before(function () {
    this.sandbox = sinon.sandbox.create();
  });

  beforeEach(function () {
    // Context stub with session set
    ctx = new Context({});
    ctx.config = new Config(process.cwd());
    ctx.daemon = new Daemon(ctx.config);
    ctx.session = new Session({ token: 'bb', passphrase: 'dd' });
    ctx.api = api.build({ auth_token: ctx.session.token });
    ctx.params = ['abc123abc'];

    ctx.target = new Target({
      path: process.cwd(),
      context: null
    });

    this.sandbox.stub(orgCreate.output, 'success');
    this.sandbox.stub(orgCreate.output, 'failure');
    this.sandbox.stub(orgCreate, '_prompt')
    .returns(Promise.resolve({
      name: ORG.body.name
    }));
    this.sandbox.stub(ctx.api.orgs, 'create').returns(Promise.resolve([ORG]));
  });

  afterEach(function () {
    this.sandbox.restore();
  });

  describe('execute', function () {
    it('errors if org is provided invalid', function () {
      ctx.params = ['@@2'];

      return orgCreate.execute(ctx).then(function () {
        assert.ok(false, 'should error');
      }, function (err) {
        assert.ok(err);
        assert.strictEqual(err.message,
          'name: Only alphanumeric, hyphens and underscores are allowed');
      });
    });

    it('prompts the user to enter an org when none is provided', function () {
      ctx.params = [];

      return orgCreate.execute(ctx).then(function () {
        sinon.assert.calledWith(orgCreate._prompt, { name: undefined });
      });
    });

    it('prompts the user to enter an org with user provided default', function () {
      ctx.params = ['default'];

      return orgCreate.execute(ctx).then(function () {
        sinon.assert.calledWith(orgCreate._prompt, { name: 'default' });
      });
    });

    it('creates an org', function () {
      return orgCreate.execute(ctx).then(function (result) {
        assert.deepEqual(result, [ORG]);
      });
    });

    it('returns error if api returns error', function () {
      ctx.api.orgs.create.returns(Promise.reject(new Error('bad')));
      return orgCreate.execute(ctx).then(function () {
        assert.ok(false, 'should error');
      }, function (err) {
        assert.ok(err);
        assert.strictEqual(err.message, 'bad');
      });
    });
  });
});
