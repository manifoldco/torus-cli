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

describe('Login', function() {
  before(function() {
    this.sandbox = sinon.sandbox.create();
  });
  describe('attempt', function() {
    beforeEach(function() {
      this.sandbox.stub(kdf, 'generate').returns(Promise.resolve(BUFFER));
      this.sandbox.stub(utils, 'hmac').returns(Promise.resolve(BUFFER));
      this.sandbox.stub(user, 'pwh').returns(PWH);
      this.sandbox.stub(client, 'post')
        .onFirstCall().returns(Promise.resolve(LOGIN_TOKEN_RESPONSE))
        .onSecondCall().returns(Promise.resolve(AUTH_TOKEN_RESPONSE));
    });
    afterEach(function() {
      this.sandbox.restore();
    });
    it('requests a loginToken from the registry', function() {
      return login.attempt({
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
      return login.attempt({
        password: PLAINTEXT,
        email: EMAIL
      }).then(function() {
        sinon.assert.calledOnce(kdf.generate);
        var salt = LOGIN_TOKEN_RESPONSE.body.salt;
        kdf.generate.calledWith(PLAINTEXT, salt);
      });
    });
    it('generates an hmac of the pwh and loginToken', function() {
      return login.attempt({
        password: PLAINTEXT,
        email: EMAIL
      }).then(function() {
        sinon.assert.calledOnce(utils.hmac);
        utils.hmac.calledWith(BUFFER);
      });
    });
    it('exchanges loginToken and pwh_hmac for authToken', function() {
      return login.attempt({
        password: PLAINTEXT,
        email: EMAIL
      }).then(function() {
        sinon.assert.calledTwice(client.post);
        var secondCall = client.post.secondCall;
        assert.deepEqual(secondCall.args[0], {
          url: '/login',
          json: {
            pwh_hmac: base64url.encode(BUFFER)
          }
        });
      });
    });
  });
});
