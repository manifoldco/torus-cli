'use strict';

var net = require('net');
var assert = require('assert');

var Promise = require('es6-promise').Promise;
var sinon = require('sinon');

var Client = require('../../lib/daemon/client');

describe('Daemon Client', function() {

  var client;
  describe('constructor', function() {

    it('throws an error on bad socketpath', function() {
      assert.throws(function() {
        client = new Client(false);
      }, /socketPath string must be provided/);
    });

    it('creates a socket and sets up subscriptions', function() {
      client = new Client('/tmp/socket');

      assert.strictEqual(client.socketPath, '/tmp/socket');
      assert.ok(client.socket instanceof net.Socket);
      assert.ok(client.subscriptions);
      assert.strictEqual(client.subscriptions.subscriptions.length, 7);
    });
  });

  describe('#connect', function() {

    beforeEach(function() {
      client = new Client('/tmp/socket');
    });

    it('propagates the error', function () {
      sinon.stub(client.socket, 'connect', function(obj, cb) {
        cb(new Error('hi'));
      });

      return client.connect().then(function() {
        assert.ok(false, 'shouldnt have succeeded');
      }).catch(function(err) {
        assert.ok(err instanceof Error);
        assert.strictEqual(err.message, 'hi');
      });
    });

    it('returns successfully', function () {
      sinon.stub(client.socket, 'connect', function(obj, cb) {
        cb();
      });

      return client.connect();
    });
  });

  describe('#send', function () {
    beforeEach(function() {
      client = new Client('/tmp/socket');
      sinon.stub(client.socket, 'write', function(data, cb) {
        cb();
      });
    });

    it('catches bad msg: missing msg', function() {
      return client.send(false).then(function() {
        assert.ok(false, 'shouldnt happen');
      }).catch(function(err) {
        assert.ok(err instanceof Error);
        assert.strictEqual(err.message, 'Message missing required properties');
      });
    });

    it('catches bad msg: missing id', function() {
      return client.send({ type: 'request' }).then(function() {
        assert.ok(false, 'shouldnt happen');
      }).catch(function(err) {
        assert.ok(err instanceof Error);
        assert.strictEqual(err.message, 'Message missing required properties');
      });
    });

    it('catches bad msg: missing type', function () {
      return client.send({ id: 'banana', command: 'request' }).then(function() {
        assert.ok(false, 'shouldnt happen')
      }).catch(function(err) {
        assert.ok(err instanceof Error);
        assert.strictEqual(err.message, 'Message missing required properties');
      });
    });

    it('catches bad msg: missing command', function () {
      return client.send({ id: 'banana', type: 'request' }).then(function() {
        assert.ok(false, 'shouldnt happen')
      }).catch(function(err) {
        assert.ok(err instanceof Error);
        assert.strictEqual(err.message, 'Message missing required properties');
      });
    });

    it('sends the msg as json', function() {
      var msg = {
        id: 'baba',
        command: 'sdf',
        type: 'request'
      };
      var msgStr = JSON.stringify(msg);

      return client.send(msg).then(function() {
        sinon.assert.calledOnce(client.socket.write);
        sinon.assert.calledWith(client.socket.write, msgStr, sinon.match.any);
      });
    });
  });

  describe('#end', function() {
    it('ends the socket', function() {
      client = new Client('/tmp/socket');
      sinon.stub(client.socket, 'end');

      client.end();

      sinon.assert.calledOnce(client.socket.end);
    });
  });

  describe('#_onData', function() {

    beforeEach(function() {
      client = new Client('/tmp/socket');
      sinon.spy(client, 'emit');
    });

    it('handles empty buffer', function() {
      client._onData(JSON.stringify({"a":"b"})+'\n');

      sinon.assert.calledOnce(client.emit);
      sinon.assert.calledWith(client.emit, 'message', { a: 'b' });
    });

    it('handles half-split messages', function() {
      client.buf = '{"a"';

      client._onData(':"b"}\n');

      sinon.assert.calledOnce(client.emit);
      sinon.assert.calledWith(client.emit, 'message', { a: 'b' });
    });

    it('pushes remainder onto buffer', function() {
      client._onData('{"a":"b"}\n{"c":');

      assert.strictEqual(client.buf, '{"c":');
    });

    it('throws an error on bad json', function() {
      var spy = sinon.spy();

      client.on('error', spy);
      client._onData('abcd\n');

      sinon.assert.calledOnce(client.emit);
      sinon.assert.calledWith(client.emit, 'error', sinon.match.any);
      sinon.assert.calledOnce(spy);

      var err = spy.getCall(0).args[0];
      assert.ok(/Could not parse message:/.test(err.message));
    });
  });

  describe('#_onTimeout', function() {
    it('destroys the socket and emits an error', function() {
      var spy = sinon.spy();

      client = new Client('/tmp/socket');
      client.on('error', spy);
      sinon.stub(client.socket, 'destroy');

      client._onTimeout();

      sinon.assert.calledOnce(spy);
      sinon.assert.calledOnce(client.socket.destroy);

      var err = spy.getCall(0).args[0];
      assert.strictEqual(err.message, 'Socket timeout');
    });
  });

  describe('#_onError', function() {
    it('destroys the socket and emits an error', function() {
      var spy = sinon.spy();

      client = new Client('/tmp/socket');
      client.on('error', spy);
      sinon.stub(client.socket, 'destroy');

      client._onError(new Error('hi'));

      sinon.assert.calledOnce(spy);
      sinon.assert.calledOnce(client.socket.destroy);

      var err = spy.getCall(0).args[0];
      assert.strictEqual(err.message, 'hi');
    });
  });
});
