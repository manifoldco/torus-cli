'use strict';

var sinon = require('sinon');
var assert = require('assert');
var Promise = require('es6-promise').Promise;
var utils = require('common/utils');

var status = require('../../lib/user/status');
var client = require('../../lib/api/client').create();
var Config = require('../../lib/config');
var Context = require('../../lib/cli/context');
var Daemon = require('../../lib/daemon/object').Daemon;

var CTX = new Context({});
CTX.config = new Config(process.cwd());
CTX.daemon = new Daemon(CTX.config);

var USER_RESPONSE = {
  body: [
    {
      id: utils.id('user'),
      body: {
        name: 'Jim Bob',
        email: 'jim@example.com',
      }
    }
  ]
};

describe('Session', function() {
  before(function() {
    this.sandbox = sinon.sandbox.create();
  });
  beforeEach(function() {
    this.sandbox.stub(status.output, 'success');
    this.sandbox.stub(status.output, 'failure');
    this.sandbox.stub(client, 'get')
      .returns(Promise.resolve(USER_RESPONSE));
  });
  afterEach(function() {
    this.sandbox.restore();
  });
  describe('execute', function() {
    beforeEach(function() {
      CTX.token = undefined;
    });

    describe('unauthenticated', function() {
      it('resolves a null user when no token present', function() {
        return status.execute(CTX).then(function(identity) {
          assert.strictEqual(identity.user, null);
        });
      });

      it('returns null user when unauthorized', function() {
        CTX.token = 'this is defined';
        client.get.restore();
        this.sandbox.stub(client, 'get')
          .returns(Promise.reject({ type: 'unauthorized' }));

        return status.execute(CTX).then(function(identity) {
          assert.strictEqual(identity.user, null);
        });
      });
    });

    describe('authenticated', function() {
      beforeEach(function() {
        CTX.token = 'this is defined';
      });

      it('calls /users/self api endpoint', function() {
        return status.execute(CTX).then(function() {
          sinon.assert.calledWith(client.get, {
            url: '/users/self'
          });
        });
      });

      it('returns null user when not found', function() {
        client.get.restore();
        this.sandbox.stub(client, 'get')
          .returns(Promise.reject({ type: 'not_found' }));

        return status.execute(CTX).then(function(identity) {
          assert.strictEqual(identity.user, null);
        });
      });
    });
  });
});
