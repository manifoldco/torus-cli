/* eslint-env mocha */

'use strict';

var sinon = require('sinon');
var assert = require('assert');
var Promise = require('es6-promise').Promise;
var utils = require('common/utils');

var Session = require('../../lib/session');
var status = require('../../lib/context/status');
var client = require('../../lib/api/client').create();
var Target = require('../../lib/context/target');
var Config = require('../../lib/config');
var Context = require('../../lib/cli/context');
var Daemon = require('../../lib/daemon/object').Daemon;

var USER_RESPONSE = {
  body: [
    {
      id: utils.id('user'),
      body: {
        name: 'Jim Bob',
        email: 'jim@example.com'
      }
    }
  ]
};

describe('Session', function () {
  before(function () {
    this.sandbox = sinon.sandbox.create();
  });

  var ctx;
  beforeEach(function () {
    ctx = new Context({});
    ctx.config = new Config(process.cwd());
    ctx.daemon = new Daemon(ctx.config);
    ctx.session = new Session({ token: 'aa', passphrase: 'safsd' });
    ctx.target = new Target({
      path: process.cwd(),
      context: {
        org: 'myorg',
        project: 'myproject'
      }
    });
    this.sandbox.stub(status.output, 'success');
    this.sandbox.stub(status.output, 'failure');
    this.sandbox.stub(client, 'get')
      .returns(Promise.resolve(USER_RESPONSE));
  });
  afterEach(function () {
    this.sandbox.restore();
  });

  describe('execute', function () {
    it('errors if target context is not provided', function () {
      ctx.target = undefined;
      return status.execute(ctx).then(function () {
        assert.ok(false, 'should not pass');
      }, function (err) {
        assert.ok(err);
        assert.strictEqual(
          err.message, 'Target must be on the context object');
      });
    });

    it('errors if api response is invalid', function () {
      client.get.onCall(0).returns(Promise.resolve({ body: null }));

      return status.execute(ctx).then(function () {
        assert.ok(false, 'should not pass');
      }, function (err) {
        assert.ok(err);
        assert.strictEqual(
          err.message, 'Invalid response returned from the API');
      });
    });

    it('returns the right properties', function () {
      return status.execute(ctx).then(function (result) {
        assert.deepEqual(result, {
          user: USER_RESPONSE.body[0],
          target: ctx.target
        });
      });
    });
  });
});
