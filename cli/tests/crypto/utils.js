'use strict';

var assert = require('assert');

var utils = require('../../lib/crypto/utils');

describe('Crypto', function() {
  describe('utils', function() {
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
