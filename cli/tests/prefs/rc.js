/* eslint-env mocha */

'use strict';

var assert = require('assert');

var Promise = require('es6-promise').Promise;
var sinon = require('sinon');

var rc = require('../../lib/prefs/rc');

var fs = require('fs');
var lock = require('../../lib/util/lock');
var Prefs = require('../../lib/prefs');

describe('rc', function () {
  var sandbox;
  var prefs;
  var stat = { isFile: function () {} };
  var rcPath = '/user/home/.arigatorc';
  var rcContents = {
    core: { context: true }
  };

  beforeEach(function () {
    prefs = new Prefs(rcPath, rcContents);
    stat.mode = 33188;

    sandbox = sinon.sandbox.create();

    sandbox.stub(stat, 'isFile').returns(true);
    sandbox.stub(fs, 'stat').callsArgWith(1, null, stat);
    sandbox.stub(lock, 'acquire').returns(Promise.resolve());
    sandbox.stub(lock, 'release').returns(Promise.resolve());
    sandbox.stub(rc, '_write').returns(Promise.resolve());
    sandbox.stub(rc, '_read').returns(Promise.resolve());
  });

  afterEach(function () {
    sandbox.restore();
  });

  describe('#stat', function () {
    it('returns false on ENOENT', function () {
      fs.stat.callsArgWith(1, { code: 'ENOENT' });

      return rc.stat(rcPath).then(function (exists) {
        assert(!exists);
      });
    });

    it('will not resolve if fs.stat fails', function () {
      fs.stat.callsArgWith(1, 'err');

      return rc.stat(rcPath).then(function () {
        assert.ok(false, 'should not resolve');
      }).catch(function (err) {
        assert.ok(err);
      });
    });

    it('will not resolve if .arigatorc is not a file', function () {
      stat.isFile.returns(false);

      return rc.stat(rcPath).then(function () {
        assert.ok(false, 'should not resolve');
      }).catch(function (err) {
        assert.ok(err);
        assert.strictEqual(err.message, '.arigatorc must be a file: ' + rcPath);
      });
    });

    it('will not resolve on invalid file permissions', function () {
      stat.mode = 0;

      return rc.stat(rcPath).then(function () {
        assert.ok(false, 'should not resolve');
      }).catch(function (err) {
        assert.ok(err);
        assert.strictEqual(
          err.message,
          'rc file permission error: /user/home/.arigatorc 00 not 0644'
        );
      });
    });

    it('will resolve with true', function () {
      return rc.stat(rcPath).then(function (exists) {
        assert.ok(exists);
      });
    });
  });

  describe('#write', function () {
    beforeEach(function () {
      sandbox.stub(rc, 'stat').returns(Promise.resolve(true));
    });

    it('will not resolve if prefs is not an instance of Prefs', function () {
      rc.write('not prefs').then(function () {
        assert.ok(false, 'should not resolve');
      }).catch(function (err) {
        assert.ok(err);
        assert.strictEqual(err.message, 'prefs must be an instance of Prefs');
      });
    });

    it('will call _write', function () {
      rc.write(prefs).then(function () {
        assert(rc._write.calledWith(rcPath, rcContents));
      });
    });
  });

  describe('#read', function () {
    beforeEach(function () {
      sandbox.stub(rc, 'stat').returns(Promise.resolve(true));
    });

    it('throws if rcPath is invalid', function () {
      assert.throws(function () {
        rc.read(null);
      }, /Must provide rcPath string/);
    });

    it('throws if rcPath is not absolute', function () {
      assert.throws(function () {
        rc.read('boo/urns/');
      }, /Must provide an absolute rc path/);
    });

    it('will call _read', function () {
      rc.read(rcPath).then(function () {
        assert(rc._read.calledWith(rcPath));
      });
    });
  });
});
