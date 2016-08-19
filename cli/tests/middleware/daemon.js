/* eslint-env mocha */

'use strict';

var sinon = require('sinon');
var assert = require('assert');
var Promise = require('es6-promise').Promise;

var Context = require('../../lib/cli/context');
var Config = require('../../lib/config');
var daemon = require('../../lib/daemon');
var api = require('../../lib/api');
var middleware = require('../../lib/middleware/daemon');

describe('daemon middleware', function () {
  var sandbox;
  var ctx;
  var mockAPI;

  beforeEach(function () {
    sandbox = sinon.sandbox.create();
    mockAPI = api.build({
      registryUrl: 'https://registry.url',
      socketUrl: 'https://socket.url'
    });
    sandbox.stub(mockAPI.versionApi, 'get').returns(Promise.resolve({
      daemon: { version: '0.0.1' }
    }));
    sandbox.stub(api, 'build').returns(mockAPI);
    ctx = new Context({});
    ctx.config = new Config(process.cwd());
    ctx.config.version = '0.0.1';
    ctx.cmd = {};
  });

  afterEach(function () {
    sandbox.restore();
  });

  it('does not try and start the daemon if already running', function () {
    sandbox.stub(daemon, 'status').returns(Promise.resolve({
      exists: true,
      pid: 100
    }));
    sandbox.stub(daemon, 'start').returns(Promise.resolve());

    return middleware.preHook()(ctx).then(function () {
      sinon.assert.notCalled(daemon.start);
    });
  });

  it('starts daemon if its not running', function () {
    sandbox.stub(daemon, 'status').returns(Promise.resolve({
      exists: false,
      pid: null
    }));
    sandbox.stub(daemon, 'start').returns(Promise.resolve());

    return middleware.preHook()(ctx).then(function () {
      sinon.assert.calledOnce(daemon.start);
    });
  });

  it('skips daemon creation if cmd group is prefs', function () {
    sandbox.spy(daemon, 'status');
    sandbox.spy(daemon, 'start');
    ctx.cmd.group = 'prefs';

    return middleware.preHook()(ctx).then(function () {
      sinon.assert.notCalled(daemon.start);
      sinon.assert.notCalled(daemon.status);
    });
  });

  it('restarts daemon if the versions do not match', function () {
    mockAPI.versionApi.get
      .returns(Promise.resolve({ daemon: { version: '0' } }))
      .onSecondCall()
      .returns(Promise.resolve({ daemon: { version: '0.0.1' } }));

    sandbox.stub(daemon, 'status').returns(Promise.resolve({
      exists: true,
      pid: 100
    }));
    sandbox.stub(daemon, 'start').returns(Promise.resolve());
    sandbox.stub(daemon, 'restart').returns(Promise.resolve(mockAPI));

    return middleware.preHook()(ctx).then(function () {
      sinon.assert.calledOnce(daemon.restart);
    });
  });

  it('errors if the versions do not match after a restart', function () {
    mockAPI.versionApi.get
      .returns(Promise.resolve({ daemon: { version: '0' } }))
      .onSecondCall()
      .returns(Promise.resolve({ daemon: { version: '0' } }));

    sandbox.stub(daemon, 'status').returns(Promise.resolve({
      exists: true,
      pid: 100
    }));
    sandbox.stub(daemon, 'start').returns(Promise.resolve());
    sandbox.stub(daemon, 'restart').returns(Promise.resolve(mockAPI));

    return middleware.preHook()(ctx).then(function () {
      sinon.assert.calledOnce(daemon.restart);
    }).catch(function (err) {
      var errMsg = 'Wrong version of daemon running, check for zombie ag-daemon process';
      assert.strictEqual(err.message, errMsg);
    });
  });
});
