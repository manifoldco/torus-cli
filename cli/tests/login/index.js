'use strict';

var sinon = require('sinon');
var assert = require('assert');
var base64url = require('base64url');
var Promise = require('es6-promise').Promise;

var kdf = require('../../lib/crypto/kdf');
var user = require('../../lib/user/crypto');
var utils = require('../../lib/crypto/utils');
var login = require('../../lib/login');
var client = require('../../lib/api/client').create();

var Config = require('../../lib/config');
var Context = require('../../lib/cli/context');
var Daemon = require('../../lib/daemon/object').Daemon;

var PLAINTEXT = 'password';
var EMAIL = 'jeff@example.com';
var BUFFER = new Buffer('buffering');
var PWH = 'this_is_a_password_hash';
var AUTH_TOKEN_RESPONSE = {
  body: {
    auth_token: 'you shall pass'
  }
};
var LOGIN_TOKEN_RESPONSE = {
  body: {
    salt: 'taffy',
    login_token: 'can I pass?',
  }
};

var CTX = new Context({});
CTX.config = new Config(process.cwd());
CTX.daemon = new Daemon(CTX.config);

describe('Login', function() {
  before(function() {
    this.sandbox = sinon.sandbox.create();
  });
  beforeEach(function() {
    this.sandbox.stub(login.output, 'success');
    this.sandbox.stub(login.output, 'failure');
    this.sandbox.stub(client, 'post')
      .onFirstCall().returns(Promise.resolve(LOGIN_TOKEN_RESPONSE))
      .onSecondCall().returns(Promise.resolve(AUTH_TOKEN_RESPONSE));
    this.sandbox.stub(CTX.daemon, 'get').returns(Promise.resolve());
    this.sandbox.stub(CTX.daemon, 'set').returns(Promise.resolve());
  });
  afterEach(function() {
    this.sandbox.restore();
  });
  describe('execute', function() {
    it('skips the prompt when inputs are supplied', function() {
      this.sandbox.stub(login, '_prompt').returns(Promise.resolve());
      this.sandbox.stub(login, '_execute').returns(Promise.resolve());
      return login.execute(CTX, { inputs: true }).then(function() {
        sinon.assert.notCalled(login._prompt);
      });
    });
    it('calls prompt.start when inputs are not supplied', function() {
      this.sandbox.stub(login, '_prompt').returns(Promise.resolve());
      this.sandbox.stub(login, '_execute').returns(Promise.resolve());
      return login.execute(CTX).then(function() {
        sinon.assert.calledOnce(login._prompt);
      });
    });
  });
  describe('subcommand', function() {
    it('calls execute with inputs', function() {
      var inputs = {
        email: 'jeff@example.com',
        passphrase: 'password',
      };

      this.sandbox.stub(login, 'execute').returns(Promise.resolve());
      return login.subcommand(CTX, inputs).then(function() {
        sinon.assert.calledOnce(login.execute);
      });
    });
    it('calls the failure output when rejecting', function(done) {
      var inputs = {};
      login.subcommand(CTX, inputs).then(function() {
        done(new Error('should not call'));
      }).catch(function() {
        sinon.assert.calledOnce(login.output.failure);
        done();
      });
    });
    it('flags err output false on rejection', function(done) {
      var inputs = {};
      login.subcommand(CTX, inputs).then(function() {
        done(new Error('should not call'));
      }).catch(function(err) {
        assert.equal(err.output, false);
        done();
      });
    });
  });
  describe('_execute', function() {
    beforeEach(function() {
      this.sandbox.stub(kdf, 'generate').returns(Promise.resolve(BUFFER));
      this.sandbox.stub(utils, 'hmac').returns(Promise.resolve(BUFFER));
      this.sandbox.stub(user, 'pwh').returns(PWH);
    });
    it('requests a loginToken from the registry', function() {
      return login._execute(CTX, {
        password: PLAINTEXT,
        email: EMAIL
      }).then(function() {
        sinon.assert.calledTwice(client.post);
        var firstCall = client.post.firstCall;
        assert.deepEqual(firstCall.args[0], {
          url: '/login/session',
          json: {
            email: EMAIL
          }
        });
      });
    });
    it('derives a high entropy password from plaintext pw', function() {
      return login._execute(CTX, {
        password: PLAINTEXT,
        email: EMAIL
      }).then(function() {
        sinon.assert.calledOnce(kdf.generate);
        var salt = LOGIN_TOKEN_RESPONSE.body.salt;
        kdf.generate.calledWith(PLAINTEXT, salt);
      });
    });
    it('generates an hmac of the pwh and loginToken', function() {
      return login._execute(CTX, {
        password: PLAINTEXT,
        email: EMAIL
      }).then(function() {
        sinon.assert.calledOnce(utils.hmac);
        utils.hmac.calledWith(BUFFER);
      });
    });
    it('exchanges loginToken and pwh_hmac for authToken', function() {
      return login._execute(CTX, {
        password: PLAINTEXT,
        email: EMAIL
      }).then(function() {
        sinon.assert.calledTwice(client.post);
        var secondCall = client.post.secondCall;
        assert.deepEqual(secondCall.args[0], {
          url: '/login',
          json: {
            login_token_hmac: base64url.encode(BUFFER)
          }
        });
      });
    });
  });
});
