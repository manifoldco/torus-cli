'use strict';

var assert = require('assert');
var crypto = require('crypto');
var sinon = require('sinon');

var utils = require('../../lib/crypto/utils');

describe('Crypto', function() {
  describe('utils', function() {
    describe('#hmac', function() {
      before(function() {
        this.sandbox = sinon.sandbox.create();
      });
      it('calls crypto.hmac with sha512', function() {
        var contents = JSON.stringify({ name: 'Yes' });
        var key = 'super secret';
        var buffer = new Buffer('buffering');
        var hmacStub = {
          update: sinon.stub(),
          digest: sinon.stub().returns(buffer)
        };
        this.sandbox.stub(crypto, 'createHmac').returns(hmacStub);
        return utils.hmac(contents, key).then(function(val) {
          assert.ok(Buffer.isBuffer(val));
          crypto.createHmac.calledWith('sha512', key);
          hmacStub.update.calledWith(contents);
          assert.strictEqual(val, buffer);
        });
      });
      it('returns buffer', function() {
        var contents = JSON.stringify({ name: 'Yes' });
        var key = 'super secret';
        return utils.hmac(contents, key).then(function(val) {
          assert.ok(Buffer.isBuffer(val));
        });
      });
    });
    describe('#randomBytes', function() {
      it('fails when length not supplied', function() {
        assert.throws(function() {
          utils.randomBytes();
        }, /length required/);
      });
      it('fails when length not greated than 1', function() {
        assert.throws(function() {
          utils.randomBytes(0);
        }, /length required/);
      });
      it('returns buffer', function() {
        return utils.randomBytes(1).then(function(val) {
          assert.ok(Buffer.isBuffer(val));
          assert.strictEqual(val.length, 1);
        });
      });
    });
  });
});
