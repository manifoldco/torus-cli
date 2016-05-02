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

describe('Crypto', function() {
  before(function() {
    this.sandbox = sinon.sandbox.create();
  });
  describe('users', function() {
    var pwBytes, mkBytes, pwCipher, mkCipher;
    beforeEach(function() {
      pwBytes = new Buffer(16);
      mkBytes = new Buffer(256);
      pwCipher = new Buffer(224);
      mkCipher = new Buffer('masterkey cipher');
      this.sandbox.stub(kdf, 'generate')
        .returns(Promise.resolve(pwCipher));
      this.sandbox.stub(utils, 'randomBytes')
        .onFirstCall().returns(Promise.resolve(pwBytes))
        .onSecondCall().returns(Promise.resolve(mkBytes));
      this.sandbox.stub(triplesec, 'encrypt')
        .returns(Promise.resolve(mkCipher));
    });
    afterEach(function() {
      this.sandbox.restore();
    });

    describe('#encryptPasswordObject', function() {
      it('generates 16byte salt for password', function() {
        return user.encryptPasswordObject(PLAINTEXT).then(function() {
          sinon.assert.calledTwice(utils.randomBytes);
          var firstCall = utils.randomBytes.firstCall;
          assert.strictEqual(firstCall.args[0], 16);
        });
      });

      it('encrypts plaintext password with generated salt', function() {
        return user.encryptPasswordObject(PLAINTEXT).then(function() {
          sinon.assert.calledOnce(kdf.generate);
          var firstCall = kdf.generate.firstCall;
          assert.strictEqual(firstCall.args[0], PLAINTEXT);
          assert.strictEqual(firstCall.args[1], pwBytes);
        });
      });

      it('generates 256byte key for master key', function() {
        return user.encryptPasswordObject(PLAINTEXT).then(function() {
          sinon.assert.calledTwice(utils.randomBytes);
          var secondCall = utils.randomBytes.secondCall;
          assert.strictEqual(secondCall.args[0], 256);
        });
      });

      it('encrypts master key with password slice', function() {
        return user.encryptPasswordObject(PLAINTEXT).then(function() {
          sinon.assert.calledOnce(triplesec.encrypt);
          var firstCall = triplesec.encrypt.firstCall;
          assert.deepEqual(firstCall.args[0], {
            data: mkBytes,
            key: pwCipher.slice(0, 192)
          });
        });
      });

      it('returns password and master objects', function() {
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

      it('fails when byte length is incorrect', function(done) {
        kdf.generate.restore();
        this.sandbox.stub(kdf, 'generate')
          .returns(Promise.resolve(new Buffer(1)));
        user.encryptPasswordObject(PLAINTEXT).then(function() {
          done(new Error('should not call'));
        }).catch(function(err) {
          assert.equal(err.message, 'invalid buffer length');
          done();
        });
      });
    });
  });
});
