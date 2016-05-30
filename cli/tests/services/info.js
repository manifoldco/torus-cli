/* eslint-env mocha */

'use strict';

var assert = require('assert');

var sinon = require('sinon');
var Promise = require('es6-promise').Promise;
var utils = require('common/utils');

var Session = require('../../lib/session');
var Context = require('../../lib/cli/context');
var ValidationError = require('../../lib/validate').ValidationError;

var client = require('../../lib/api/client').create();
var serviceInfo = require('../../lib/services/info');

var ORG = {
  id: utils.id('org'),
  body: {
    name: 'my-org'
  }
};

var SERVICE = {
  id: utils.id('service'),
  body: {
    name: 'api-1',
    org_id: ORG.id
  }
};


describe('Services Info', function () {
  var sandbox;
  var ctx;

  beforeEach(function () {
    sandbox = sinon.sandbox.create();
    ctx = new Context({});
    ctx.session = new Session({
      token: 'abcdefgh',
      passphrase: 'hohohoho'
    });
    ctx.params = ['hi-there'];
  });

  afterEach(function () {
    sandbox.restore();
  });

  describe('#execute', function () {
    it('rejects a invalid service name', function () {
      ctx.params = ['!!!!!@#'];

      return serviceInfo.execute(ctx).then(function () {
        assert.ok(false, 'should not resolve');
      }).catch(function (err) {
        assert.ok(err);
        assert.ok(err instanceof ValidationError);
      });
    });

    it('requires an auth token', function () {
      ctx.session = null;

      return serviceInfo.execute(ctx).then(function () {
        assert.ok(false, 'should not resolve');
      }).catch(function (err) {
        assert.ok(err);
        assert.strictEqual(err.message, 'Session object missing on Context');
      });
    });

    it('returns not found', function () {
      sandbox.stub(client, 'get').returns(Promise.resolve({ body: [] }));

      return serviceInfo.execute(ctx).then(function () {
        assert.ok(false, 'should not resolve');
      }).catch(function (err) {
        assert.ok(err);
        assert.strictEqual(err.message, 'service not found');
      });
    });

    it('handles not found from api', function () {
      var err = new Error('service not found');
      err.type = 'not_found';
      sandbox.stub(client, 'get').returns(Promise.reject(err));

      return serviceInfo.execute(ctx).then(function () {
        assert.ok(false, 'should not resolve');
      }).catch(function (e) {
        assert.strictEqual(e.type, 'not_found');
      });
    });

    it('resolves with a service object', function () {
      sandbox.stub(client, 'get').returns(Promise.resolve({
        body: [SERVICE]
      }));

      return serviceInfo.execute(ctx).then(function (service) {
        assert.deepEqual(service, SERVICE);
      });
    });
  });
});
