/* eslint-env mocha */

'use strict';

var uuid = require('uuid');
var sinon = require('sinon');
var assert = require('assert');
var Promise = require('es6-promise').Promise;

var verify = require('../../lib/verify');
var Session = require('../../lib/session');
var api = require('../../lib/api');
var Config = require('../../lib/config');
var Context = require('../../lib/cli/context');
var Daemon = require('../../lib/daemon/object').Daemon;

var USER = {
  id: uuid.v4(),
  body: {
    name: 'Jimbob',
    email: 'jim@example.com'
  }
};

var VERIFY_RESPONSE = {
  user: USER
};

describe('Verify', function () {
  var ctx;
  before(function () {
    this.sandbox = sinon.sandbox.create();
  });
  beforeEach(function () {
    ctx = new Context({});
    ctx.config = new Config(process.cwd());
    ctx.daemon = new Daemon(ctx.config);
    ctx.params = ['ABC123ABC'];
    ctx.session = new Session({ token: 'aa', passphrase: 'dd' });
    ctx.api = api.build({ auth_token: ctx.session.token });

    this.sandbox.stub(verify.output, 'success');
    this.sandbox.stub(verify.output, 'failure');
    this.sandbox.stub(ctx.api.users, 'verify')
      .returns(Promise.resolve(VERIFY_RESPONSE));
  });
  afterEach(function () {
    this.sandbox.restore();
  });
  describe('execute', function () {
    it('calls _execute with inputs', function () {
      this.sandbox.stub(verify, '_prompt').returns(Promise.resolve());
      this.sandbox.stub(verify, '_execute').returns(Promise.resolve());
      return verify.execute(ctx).then(function () {
        sinon.assert.calledOnce(verify._execute);
      });
    });
    it('skips the prompt when inputs are supplied', function () {
      this.sandbox.stub(verify, '_prompt').returns(Promise.resolve());
      this.sandbox.stub(verify, '_execute').returns(Promise.resolve());
      return verify.execute(ctx).then(function () {
        sinon.assert.notCalled(verify._prompt);
      });
    });
  });
  describe('subcommand', function () {
    it('calls execute', function () {
      this.sandbox.stub(verify, 'execute').returns(Promise.resolve());
      return verify.subcommand(ctx).then(function () {
        sinon.assert.calledOnce(verify.execute);
      });
    });
    it('calls the failure output when rejecting', function (done) {
      var err = new Error('fake err');
      this.sandbox.stub(verify, 'execute').returns(Promise.reject(err));
      verify.subcommand(ctx).then(function () {
        done(new Error('dont call'));
      }).catch(function () {
        sinon.assert.calledOnce(verify.output.failure);
        done();
      });
    });
    it('flags err output false on rejection', function (done) {
      var err = new Error('fake err');
      this.sandbox.stub(verify, 'execute').returns(Promise.reject(err));
      verify.subcommand(ctx).then(function () {
        done(new Error('dont call'));
      }).catch(function (e) {
        assert.equal(e.output, false);
        done();
      });
    });
  });
  describe('_execute', function () {
    it('sends api request to verify', function () {
      var input = { code: 'ABC123ABC' };
      return verify._execute(ctx.api, input).then(function () {
        sinon.assert.calledOnce(ctx.api.users.verify);
        var firstCall = ctx.api.users.verify.firstCall;
        var args = firstCall.args;
        assert.deepEqual(args[0], {
          code: 'ABC123ABC'
        });
      });
    });
  });
});
