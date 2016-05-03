'use strict';

var uuid = require('uuid');
var sinon = require('sinon');
var assert = require('assert');
var Promise = require('es6-promise').Promise;

var verify = require('../../lib/verify');
var client = require('../../lib/api/client').create();

var USER = {
  id: uuid.v4(),
  body: {
    name: 'Jimbob',
    email: 'jim@example.com'
  }
};

var VERIFY_RESPONSE = {
  user: USER
};

var DAEMON = {
  set: sinon.stub().returns(Promise.resolve()),
  get: sinon.stub().returns(Promise.resolve({ token: 'this is a token' })),
};

var DAEMON_EMPTY = {
  set: sinon.stub().returns(Promise.resolve()),
  get: sinon.stub().returns(Promise.resolve({ token: '' })),
};

var CTX = {
  daemon: DAEMON,
  params: ['ABC123ABC']
};

describe('Verify', function() {
  before(function() {
    this.sandbox = sinon.sandbox.create();
  });
  beforeEach(function() {
    this.sandbox.stub(verify.output, 'success');
    this.sandbox.stub(verify.output, 'failure');
    this.sandbox.stub(client, 'post')
      .returns(Promise.resolve(VERIFY_RESPONSE));
    this.sandbox.spy(client, 'auth');
  });
  afterEach(function() {
    this.sandbox.restore();
  });
  describe('execute', function() {
    it('calls _execute with inputs', function() {
      this.sandbox.stub(verify, '_prompt').returns(Promise.resolve());
      this.sandbox.stub(verify, '_execute').returns(Promise.resolve());
      return verify.execute(CTX).then(function() {
        sinon.assert.calledOnce(verify._execute);
      });
    });
    it('skips the prompt when inputs are supplied', function() {
      this.sandbox.stub(verify, '_prompt').returns(Promise.resolve());
      this.sandbox.stub(verify, '_execute').returns(Promise.resolve());
      return verify.execute(CTX).then(function() {
        sinon.assert.notCalled(verify._prompt);
      });
    });
  });
  describe('subcommand', function() {
    it('calls execute', function() {
      this.sandbox.stub(verify, 'execute').returns(Promise.resolve());
      return verify.subcommand(CTX).then(function() {
        sinon.assert.calledOnce(verify.execute);
      });
    });
    it('calls the failure output when rejecting', function(done) {
      var err = new Error('fake err');
      this.sandbox.stub(verify, 'execute').returns(Promise.reject(err));
      verify.subcommand(CTX).then(function() {
        done(new Error('dont call'));
      }).catch(function() {
        sinon.assert.calledOnce(verify.output.failure);
        done();
      });
    });
    it('flags err output false on rejection', function(done) {
      var err = new Error('fake err');
      this.sandbox.stub(verify, 'execute').returns(Promise.reject(err));
      verify.subcommand(CTX).then(function() {
        done(new Error('dont call'));
      }).catch(function(err) {
        assert.equal(err.output, false);
        done();
      });
    });
  });
  describe('_execute', function() {
    it('authorizes the client', function() {
      var input = { code: 'ABC123ABC' };
      return verify._execute(DAEMON, input).then(function() {
        sinon.assert.calledOnce(client.auth);
      });
    });
    it('fails if token not found in daemon', function(done) {
      var input = { code: 'ABC123ABC' };
      verify._execute(DAEMON_EMPTY, input).then(function() {
        done(new Error('dont call'));
      }).catch(function(err) {
        assert.equal(err.message, 'must authenticate first');
        done();
      });
    });
    it('sends api request to verify', function() {
      var input = { code: 'ABC123ABC' };
      return verify._execute(DAEMON, input).then(function() {
        sinon.assert.calledOnce(client.post);
        var firstCall = client.post.firstCall;
        var args = firstCall.args;
        assert.deepEqual(args[0], {
          url: '/users/verify',
          json: {
            code: 'ABC123ABC'
          }
        });
      });
    });
  });
});
