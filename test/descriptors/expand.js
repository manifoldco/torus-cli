'use strict';

var assert = require('assert');
var path = require('path');
var fs = require('fs');

var sinon = require('sinon');
var expand = require('../../lib/descriptors/expand');

const FILE_PATH = path.join(__dirname, '../data/ref.json');

describe('expand', function() {
  var sandbox;

  beforeEach(() => {
    sandbox = sinon.sandbox.create();
  });
  afterEach(() => {
    sandbox.restore();
  });

  it('loads the file properly', function() {
    return expand(FILE_PATH).then((data) => {
      assert.deepEqual(data, {
        a: 'b',
        c: { d: true },
        d: [
          'a',
          'b',
          { d: true },
          { b: 1, d: { d: true } }
        ]
      });
    });
  });

  it('returns an error if non-json data', function() {
    sandbox.stub(fs, 'readFile').yields(null, '\hi');
    
    return expand(FILE_PATH).then(() => {
      assert.ok(false, 'shouldnt have succeeded');
    }, (err) => {
      assert.ok(err instanceof Error);
      assert.strictEqual(err.message, 'Unexpected token h');
    });
  });

  it('properly returns on an error', function() {
    sandbox.stub(fs, 'readFile').yields(new Error('hi'));

    return expand(FILE_PATH).then(() => {
      assert.ok(false, 'shouldnt have succeeded');
    }, (err) => {
      assert.ok(err instanceof Error);
      assert.strictEqual(err.message, 'hi');
    });
  });
});
