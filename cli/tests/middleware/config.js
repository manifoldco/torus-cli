/* eslint-env mocha */

'use strict';

var fs = require('fs');

var sinon = require('sinon');
var assert = require('assert');

var config = require('../../lib/middleware/config');
var Context = require('../../lib/cli/context');

describe('config middleware', function () {
  var ctx;
  var sandbox;

  beforeEach(function () {
    ctx = new Context({ version: 'banana' });
    sandbox = sinon.sandbox.create();
  });
  afterEach(function () {
    sandbox.restore();
  });

  it('creates folder if does not exist', function () {
    var err = new Error('hi');
    err.code = 'ENOENT';

    sandbox.stub(fs, 'stat').yields(err);
    sandbox.stub(fs, 'mkdir').yields();

    return config('/tmp/folder')(ctx).then(function () {
      sinon.assert.calledOnce(fs.stat);
      sinon.assert.calledWith(fs.stat, '/tmp/folder', sinon.match.any);

      sinon.assert.calledOnce(fs.mkdir);
      sinon.assert.calledWith(fs.mkdir, '/tmp/folder', 448, sinon.match.any);

      assert.strictEqual(ctx.config.arigatoRoot, '/tmp/folder');
      assert.strictEqual(ctx.config.socketUrl,
                         'http://unix:/tmp/folder/daemon.socket:');
      assert.strictEqual(ctx.config.pidPath, '/tmp/folder/daemon.pid');
      assert.strictEqual(ctx.config.version, 'banana');
    });
  });

  it('errors if create fails', function () {
    var err = new Error('hi');
    err.code = 'ENOENT';

    sandbox.stub(fs, 'stat').yields(err);
    sandbox.stub(fs, 'mkdir').yields(new Error('hi'));

    return config('/tmp/folder')(ctx).then(function () {
      assert.ok(false, 'shouldnt happen');
    }).catch(function (e) {
      assert.ok(e instanceof Error);
      assert.strictEqual(err.message, 'hi');
    });
  });

  it('resolves if folder exists', function () {
    sandbox.stub(fs, 'stat').yields(null, {
      isDirectory: function () { return true; },
      mode: 16832 // 0700
    });
    sandbox.stub(fs, 'mkdir').yields();

    return config('/tmp/folder')(ctx).then(function () {
      sinon.assert.calledOnce(fs.stat);
      sinon.assert.calledWith(fs.stat, '/tmp/folder', sinon.match.any);
      sinon.assert.notCalled(fs.mkdir);
    });
  });

  it('errors if folder exists and bad permissions', function () {
    sandbox.stub(fs, 'stat').yields(null, {
      isDirectory: function () { return true; },
      mode: 16882 // 0700
    });
    sandbox.stub(fs, 'mkdir').yields();

    return config('/tmp/folder')(ctx).then(function () {
      assert.ok(false, 'shouldnt happen');
    }).catch(function (err) {
      assert.ok(err instanceof Error);
      assert.ok(err.message.match(/Arigato root file permission/));
    });
  });

  it('errors if exists but is not a directory', function () {
    sandbox.stub(fs, 'stat').yields(null, {
      isDirectory: function () { return false; },
      mode: 16832 // 0700
    });
    sandbox.stub(fs, 'mkdir').yields();

    return config('/tmp/folder')(ctx).then(function () {
      assert.ok(false, 'shouldnt happen');
    }).catch(function (err) {
      assert.ok(err instanceof Error);
      assert.ok(err.message.match(/Arigato Root must be a directory/));
    });
  });
});
