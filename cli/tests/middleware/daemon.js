/* eslint-env mocha */

'use strict';

var sinon = require('sinon');
var Promise = require('es6-promise').Promise;

var Context = require('../../lib/cli/context');
var Config = require('../../lib/config');
var daemon = require('../../lib/daemon');
var middleware = require('../../lib/middleware/daemon');

describe('daemon middleware', function () {
  var sandbox;
  var ctx;

  beforeEach(function () {
    sandbox = sinon.sandbox.create();
    ctx = new Context({});
    ctx.config = new Config(process.cwd());
  });
  afterEach(function () {
    sandbox.restore();
  });

  describe('preHook', function () {
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
  });
});
