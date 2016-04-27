'use strict';

var sinon = require('sinon');
var assert = require('assert');
var Promise = require('es6-promise').Promise;
var base64url = require('base64url');

var kdf = require('../../lib/crypto/kdf');
var user = require('../../lib/user/crypto');
var utils = require('../../lib/crypto/utils');
var triplesec = require('../../lib/crypto/triplesec');

var PLAINTEXT = 'password';
var BUFFER = new Buffer('buffering');

describe('Crypto', function() {
  before(function() {
    this.sandbox = sinon.sandbox.create();
  });
  describe('users', function() {
    afterEach(function() {
      this.sandbox.restore();
    });

    describe('#encryptPasswordObject', function() {
      it('generates 16byte salt for password', function() {
        this.sandbox.stub(kdf, 'generate')
          .returns(Promise.resolve(BUFFER));
        this.sandbox.stub(utils, 'randomBytes')
          .returns(Promise.resolve(BUFFER));
        this.sandbox.stub(triplesec, 'encrypt')
          .returns(Promise.resolve(BUFFER));
        return user.encryptPasswordObject(PLAINTEXT).then(function() {
          sinon.assert.calledTwice(utils.randomBytes);
          var firstCall = utils.randomBytes.firstCall;
          assert.strictEqual(firstCall.args[0], 16);
        });
      });

      it('encrypts plaintext password with generated salt', function() {
        var value = new Buffer('this is the value');
        this.sandbox.stub(kdf, 'generate')
          .returns(Promise.resolve(BUFFER));
        this.sandbox.stub(utils, 'randomBytes')
          .returns(Promise.resolve(value));
        this.sandbox.stub(triplesec, 'encrypt')
          .returns(Promise.resolve(BUFFER));
        return user.encryptPasswordObject(PLAINTEXT).then(function() {
          sinon.assert.calledOnce(kdf.generate);
          var firstCall = kdf.generate.firstCall;
          assert.strictEqual(firstCall.args[0], PLAINTEXT);
          assert.strictEqual(firstCall.args[1], value);
        });
      });

      it('generates 256byte key for master key', function() {
        var value = new Buffer('this is the value');
        this.sandbox.stub(kdf, 'generate')
          .returns(Promise.resolve(BUFFER));
        this.sandbox.stub(utils, 'randomBytes')
          .returns(Promise.resolve(value));
        this.sandbox.stub(triplesec, 'encrypt')
          .returns(Promise.resolve(BUFFER));
        return user.encryptPasswordObject(PLAINTEXT).then(function() {
          sinon.assert.calledTwice(utils.randomBytes);
          var secondCall = utils.randomBytes.secondCall;
          assert.strictEqual(secondCall.args[0], 256);
        });
      });

      it('encrypts master key with password slice', function() {
        var value = new Buffer('this is the value');
        this.sandbox.stub(kdf, 'generate')
          .returns(Promise.resolve(BUFFER));
        this.sandbox.stub(utils, 'randomBytes')
          .returns(Promise.resolve(value));
        this.sandbox.stub(triplesec, 'encrypt')
          .returns(Promise.resolve(BUFFER));
        return user.encryptPasswordObject(PLAINTEXT).then(function() {
          sinon.assert.calledOnce(triplesec.encrypt);
          var firstCall = triplesec.encrypt.firstCall;
          assert.deepEqual(firstCall.args[0], {
            data: value,
            key: BUFFER.slice(0, 192)
          });
        });
      });

      it('returns password and master objects', function() {
        var pwBytes = new Buffer('this is the password');
        var mkBytes = new Buffer('this is master key');
        var pwCipher = new Buffer('password cipher');
        var mkCipher = new Buffer('masterkey cipher');
        this.sandbox.stub(kdf, 'generate')
          .returns(Promise.resolve(pwCipher));
        this.sandbox.stub(utils, 'randomBytes')
          .onFirstCall().returns(Promise.resolve(pwBytes))
          .onSecondCall().returns(Promise.resolve(mkBytes));
        this.sandbox.stub(triplesec, 'encrypt')
          .returns(Promise.resolve(mkCipher));

        return user.encryptPasswordObject(PLAINTEXT).then(function(obj) {
          assert.deepEqual(obj, {
            password: {
              salt: base64url.encode(pwBytes),
              value: base64url.encode(pwCipher.slice(192)),
              alg: '0x23',
            },
            master: {
              alg: '0x22',
              value: base64url.encode(mkCipher)
            }
          });
        });
      });
    });
  });
});
