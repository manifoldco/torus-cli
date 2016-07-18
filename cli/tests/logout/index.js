/* eslint-env mocha */

'use strict';

var sinon = require('sinon');
var Promise = require('es6-promise').Promise;

var logout = require('../../lib/logout');

var Config = require('../../lib/config');
var Context = require('../../lib/cli/context');
var api = require('../../lib/api');

describe('Logout', function () {
  var ctx;
  before(function () {
    this.sandbox = sinon.sandbox.create();
  });

  beforeEach(function () {
    ctx = new Context({});

    ctx.config = new Config(process.cwd());
    ctx.api = api.build();

    this.sandbox.stub(ctx.api.logout, 'post').returns(Promise.resolve());
  });

  afterEach(function () {
    this.sandbox.restore();
  });

  it('sends a logout request to the registry and daemon', function () {
    return logout(ctx).then(function () {
      sinon.assert.calledOnce(ctx.api.logout.post);
    });
  });

  it('reports an error on client failure', function () {
    ctx.api.logout.post.returns(Promise.reject(new Error('hi')));

    return logout(ctx).catch(function () {
      sinon.assert.calledOnce(ctx.api.logout.post);
    });
  });
});
