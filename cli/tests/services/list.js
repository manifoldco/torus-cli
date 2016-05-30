/* eslint-env mocha */

'use strict';

var assert = require('assert');

var sinon = require('sinon');
var Promise = require('es6-promise').Promise;
var utils = require('common/utils');

var Session = require('../../lib/session');
var Context = require('../../lib/cli/context');

var client = require('../../lib/api/client').create();
var serviceList = require('../../lib/services/list');

var ORG = {
  id: utils.id('org'),
  body: {
    name: 'my-org'
  }
};

var SERVICES = [
  {
    id: utils.id('service'),
    body: {
      name: 'api-1',
      org_id: ORG.id
    }
  },
  {
    id: utils.id('service'),
    body: {
      name: 'api-2',
      org_id: ORG.id
    }
  }
];


describe('Services List', function () {
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
    it('requires an auth token', function () {
      ctx.session = null;

      return serviceList.execute(ctx).then(function () {
        assert.ok(false, 'should not resolve');
      }).catch(function (err) {
        assert.ok(err);
        assert.strictEqual(err.message, 'Session object not on Context');
      });
    });

    it('does not throw if the user has no services', function () {
      sandbox.stub(client, 'get').returns(Promise.resolve({ body: [] }));

      return serviceList.execute(ctx).then(function (services) {
        assert.deepEqual(services.body, []);
      }).catch(function () {
        assert.ok(false, 'should not resolve');
      });
    });

    it('handles not found from api', function () {
      var err = new Error('service not found');
      err.type = 'not_found';
      sandbox.stub(client, 'get').returns(Promise.reject(err));

      return serviceList.execute(ctx).then(function () {
        assert.ok(false, 'should not resolve');
      }).catch(function (e) {
        assert.strictEqual(e.type, 'not_found');
      });
    });

    it('resolves with an array of services', function () {
      sandbox.stub(client, 'get').returns(Promise.resolve(SERVICES));

      return serviceList.execute(ctx).then(function (service) {
        assert.deepEqual(service, SERVICES);
      });
    });
  });
});
