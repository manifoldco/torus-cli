/* eslint-env mocha */

'use strict';

var sinon = require('sinon');
var assert = require('assert');
var utils = require('common/utils');
var Promise = require('es6-promise').Promise;

var orgsList = require('../../lib/orgs/list');
var Session = require('../../lib/session');
var Config = require('../../lib/config');
var Context = require('../../lib/cli/context');
var Target = require('../../lib/context/target');
var Daemon = require('../../lib/daemon/object').Daemon;
var api = require('../../lib/api');

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
    // Context stub with session set
    ctx = new Context({});
    ctx.config = new Config(process.cwd());
    ctx.session = new Session({ token: 'aa', passphrase: 'boo' });
    ctx.daemon = new Daemon(ctx.config);
    ctx.params = [];
    ctx.options = {
      org: { value: ORG.body.name }
    };
    ctx.target = new Target({
      path: process.cwd(),
      context: null
    });
    ctx.api = api.build({ auth_token: ctx.session.token });

    this.sandbox.stub(orgsList.output, 'success');
    this.sandbox.stub(orgsList.output, 'failure');
    this.sandbox.stub(ctx.api.users, 'self').returns(Promise.resolve([SELF]));
    this.sandbox.stub(ctx.api.orgs, 'get').returns(Promise.resolve([ORG]));
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

    it('errors if the org was not found', function () {
      ctx.api.users.self.returns(Promise.resolve([]));
      ctx.api.orgs.get.returns(Promise.resolve([]));
      return orgsList.execute(ctx).then(function () {
        assert.ok(false, 'should error');
      }, function (err) {
        assert.ok(err);
        assert.strictEqual(err.message, 'No orgs found');
      });
    });
  });
});
