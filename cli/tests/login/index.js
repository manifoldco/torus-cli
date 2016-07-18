/* eslint-env mocha */

'use strict';

var sinon = require('sinon');
var assert = require('assert');
var Promise = require('es6-promise').Promise;

var login = require('../../lib/login');

var api = require('../../lib/api');
var Config = require('../../lib/config');
var Context = require('../../lib/cli/context');

var PLAINTEXT = 'password';
var EMAIL = 'jeff@example.com';

var ctx;
describe('Login', function () {
  before(function () {
    this.sandbox = sinon.sandbox.create();
  });
  beforeEach(function () {
    ctx = new Context({});
    ctx.config = new Config(process.cwd());
    ctx.api = api.build();

    this.sandbox.stub(login.output, 'success');
    this.sandbox.stub(login.output, 'failure');
  });
  afterEach(function () {
    this.sandbox.restore();
  });
  describe('execute', function () {
    it('skips the prompt when inputs are supplied', function () {
      this.sandbox.stub(login, '_prompt').returns(Promise.resolve());
      this.sandbox.stub(login, '_execute').returns(Promise.resolve());
      return login.execute(ctx, { inputs: true }).then(function () {
        sinon.assert.notCalled(login._prompt);
      });
    });
    it('calls prompt.start when inputs are not supplied', function () {
      this.sandbox.stub(login, '_prompt').returns(Promise.resolve());
      this.sandbox.stub(login, '_execute').returns(Promise.resolve());
      return login.execute(ctx).then(function () {
        sinon.assert.calledOnce(login._prompt);
      });
    });
  });
  describe('subcommand', function () {
    it('calls execute with inputs', function () {
      var inputs = {
        username: 'jeff',
        email: 'jeff@example.com',
        passphrase: 'password'
      };

      this.sandbox.stub(login, 'execute').returns(Promise.resolve());
      return login.subcommand(ctx, inputs).then(function () {
        sinon.assert.calledOnce(login.execute);
      });
    });
    it('calls the failure output when rejecting', function (done) {
      var inputs = {};
      login.subcommand(ctx, inputs).then(function () {
        done(new Error('should not call'));
      }).catch(function () {
        sinon.assert.calledOnce(login.output.failure);
        done();
      });
    });
    it('flags err output false on rejection', function (done) {
      var inputs = {};
      login.subcommand(ctx, inputs).then(function () {
        done(new Error('should not call'));
      }).catch(function (err) {
        assert.equal(err.output, false);
        done();
      });
    });
  });
  describe('_execute', function () {
    beforeEach(function () {
      this.sandbox.stub(ctx.api.login, 'post').returns(Promise.resolve());
    });
    it('calls login on the daemon', function () {
      return login._execute(ctx, {
        passphrase: PLAINTEXT,
        email: EMAIL
      }).then(function () {
        sinon.assert.calledOnce(ctx.api.login.post);
      });
    });
  });
});
