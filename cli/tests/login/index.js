/* eslint-env mocha */

'use strict';

var sinon = require('sinon');
var assert = require('assert');
var base64url = require('base64url');
var Promise = require('es6-promise').Promise;
var user = require('common/crypto/user');

var login = require('../../lib/login');

var api = require('../../lib/api');
var Config = require('../../lib/config');
var Context = require('../../lib/cli/context');
var Daemon = require('../../lib/daemon/object').Daemon;
var Session = require('../../lib/session');

var PLAINTEXT = 'password';
var EMAIL = 'jeff@example.com';
var BUFFER = new Buffer('buffering');
var AUTH_TOKEN_RESPONSE = { auth_token: 'you shall pass' };
var LOGIN_TOKEN_RESPONSE = { salt: 'taffy', login_token: 'can I pass?' };
var TYPE_LOGIN = 'login';
var TYPE_AUTH = 'auth';

var ctx;
describe('Login', function () {
  before(function () {
    this.sandbox = sinon.sandbox.create();
  });
  beforeEach(function () {
    ctx = new Context({});
    ctx.config = new Config(process.cwd());
    ctx.daemon = new Daemon(ctx.config);
    ctx.session = new Session({ token: 'aa', passphrase: 'aa' });
    ctx.api = api.build({ auth_token: ctx.session.token });

    this.sandbox.stub(login.output, 'success');
    this.sandbox.stub(login.output, 'failure');
    this.sandbox.stub(ctx.api.tokens, 'create')
      .onFirstCall()
      .returns(Promise.resolve(LOGIN_TOKEN_RESPONSE))
      .onSecondCall()
      .returns(Promise.resolve(AUTH_TOKEN_RESPONSE));
    this.sandbox.stub(ctx.daemon, 'get').returns(Promise.resolve());
    this.sandbox.stub(ctx.daemon, 'set').returns(Promise.resolve());
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
      this.sandbox.stub(user, 'deriveLoginHmac')
        .returns(Promise.resolve(base64url.encode(BUFFER)));
    });
    it('requests a loginToken from the registry', function () {
      return login._execute(ctx, {
        passphrase: PLAINTEXT,
        email: EMAIL
      }).then(function () {
        sinon.assert.calledTwice(ctx.api.tokens.create);
        var firstCall = ctx.api.tokens.create.firstCall;
        assert.deepEqual(firstCall.args[0], {
          type: TYPE_LOGIN,
          email: EMAIL
        });
      });
    });
    it('exchanges loginToken and pwh_hmac for authToken', function () {
      return login._execute(ctx, {
        passphrase: PLAINTEXT,
        email: EMAIL
      }).then(function () {
        sinon.assert.calledTwice(ctx.api.tokens.create);
        var secondCall = ctx.api.tokens.create.secondCall;
        assert.deepEqual(secondCall.args[0], {
          type: TYPE_AUTH,
          login_token_hmac: base64url.encode(BUFFER)
        });
      });
    });
  });
});
