'use strict';

const assert = require('assert');
const fs = require('fs');

const sinon = require('sinon');
const Descriptor = require('../../lib/descriptors/descriptor').Descriptor;

const SCHEMA_PATH = '../test/data/schema';
const FILE_PATH = '/tmp/testdata.yml';


describe('Descriptor', function() {

  var descriptor;
  var sandbox;
  beforeEach(function() {
    sandbox = sinon.sandbox.create();
  });

  afterEach(function() {
    sandbox.restore();
  });

  describe('#constructor', function() {
    it('throws an error for improper params', function() {
      assert.throws(function() {
        descriptor = new Descriptor('sdfsf', 13, 'string');
      }, TypeError);
    });

    it('sets the schema, path and contents', function() {
      descriptor = new Descriptor(SCHEMA_PATH, FILE_PATH, {});

      assert.strictEqual(descriptor.schema, SCHEMA_PATH);
      assert.strictEqual(descriptor.path, FILE_PATH);
      assert.deepEqual(descriptor.contents, {});
    });
  });

  describe('#get', function() {
    it('returns the data if it exists', function() {
      descriptor = new Descriptor(SCHEMA_PATH, FILE_PATH, {
        a: 2
      });

      assert.strictEqual(descriptor.get('a'), 2);
    });

    it('returns data at a deep path, that exists', function() {
      descriptor = new Descriptor(SCHEMA_PATH, FILE_PATH, {
        a: { b: { c: 1 } }
      });

      assert.strictEqual(descriptor.get('a.b.c.'), 1);
    });

    it('returns segment type if its an object an undefined', function() {
      descriptor = new Descriptor(SCHEMA_PATH, FILE_PATH, {
        name: 'hi'
      });

      assert.deepEqual(descriptor.get('obj'), {});
    });

    it('returns segment type if its an array and undefined', function() {
      descriptor = new Descriptor(SCHEMA_PATH, FILE_PATH, {});

      assert.deepEqual(descriptor.get('list'), []);
    });

    it('returns null if its a primitive', function() {
      descriptor = new Descriptor(SCHEMA_PATH, FILE_PATH, {});

      assert.strictEqual(descriptor.get('name'), null);
    });

    it('throws a ReferenceError if property is unknown', function() {
      descriptor = new Descriptor(SCHEMA_PATH, FILE_PATH, {});

      assert.throws(function() {
        descriptor.get('a.b.c.d');
      }, ReferenceError);
    });

    it('doesn\'t support getting deep properties that are not set', function() {
      descriptor = new Descriptor(SCHEMA_PATH, FILE_PATH, {});

      assert.throws(function() {
        descriptor.get('obj.a');
      }, ReferenceError);
    });
  });

  describe('#set', function() {
    beforeEach(function() {
      descriptor = new Descriptor(SCHEMA_PATH, FILE_PATH, {});
    });

    it('sets the value and returns if valid', function() {
      return descriptor.set('name', 'bcd').then((data) => {
        assert.strictEqual(data.name, 'bcd');
      });
    });

    it('takes in a map and sets', function () {
      return descriptor.set({ name: 'hii', 'obj.a': 'boo' }).then((data) => {
        assert.deepEqual(data, {
          name: 'hii',
          obj: {
            a: 'boo'
          }
        });
      });
    });

    it('returns error if invalid', function() {
      return descriptor.set('name', 1).catch((errors) => {
        assert.ok(Array.isArray(errors));
        assert.strictEqual(errors.length, 1);

        assert.ok(errors[0] instanceof Error);
      });
    });

    it('returns clone of data', function() {
      return descriptor.set('name', 'abcdef').then((data) => {
        data.name = 'x';
        assert.strictEqual(descriptor.get('name'), 'abcdef');
      });
    });
  });

  describe('#validate', function() {
    beforeEach(function() {
      descriptor = new Descriptor(SCHEMA_PATH, FILE_PATH, {
        name: 'woo'
      });
    });

    it('returns success', function() {
      return descriptor.validate();
    });

    it('returns errors', function() {
      descriptor.contents.name = 'y';
      return descriptor.validate().catch((errors) => {
        assert.ok(Array.isArray(errors));
        assert.strictEqual(errors.length, 1);

        assert.ok(errors[0] instanceof Error);
      });
    });
  });

  describe('#write', function() {

    beforeEach(function() {
      descriptor = new Descriptor(SCHEMA_PATH, FILE_PATH, {
        name: 'woo'
      });
    });

    it('writes yaml to file', function() {
      sandbox.stub(fs, 'writeFile').yields();

      return descriptor.write().then(() => {
        sinon.assert.calledOnce(fs.writeFile);
        sinon.assert.calledWith(fs.writeFile, FILE_PATH, 'name: woo\n');
      });
    });

    it('handles error', function() {
      sandbox.stub(fs, 'writeFile').yields(new Error('hi'));

      return descriptor.write().catch((err) => {
        assert.ok(err instanceof Error);
        assert.strictEqual(err.message, 'hi');
      });
    });
  });

  describe('.create', function() {

    it('writes an obj to disk', function() {
      sandbox.stub(fs, 'writeFile').yields();

      return Descriptor.create(Descriptor, SCHEMA_PATH, FILE_PATH, {
        name: 'abc' }).then(() => {
        sinon.assert.calledOnce(fs.writeFile);
        sinon.assert.calledWith(fs.writeFile, FILE_PATH, 'name: abc\n');
      });
    });

    it('returns error if create fails', function() {
      sandbox.stub(fs, 'writeFile').yields(new Error('hi'));

      return Descriptor.create(Descriptor, SCHEMA_PATH, FILE_PATH, {})
      .catch((err) => {
        assert.ok(err instanceof Error);
        assert.strictEqual(err.message, 'hi');
      });
    });
  });

  describe('.read', function() {

    it('reads valid data', function() {
      sandbox.stub(fs, 'readFile').yields(null, 'name: woo\n');

      return Descriptor.read(Descriptor, SCHEMA_PATH, FILE_PATH).then((obj) => {
        assert.deepEqual(obj.contents,  {
          name: 'woo'
        });

        sinon.assert.calledOnce(fs.readFile);
        sinon.assert.calledWith(fs.readFile, FILE_PATH, {
          encoding: 'utf-8'
        });
      });
    });

    it('returns error if data doesnt match schema', function() {
      sandbox.stub(fs, 'readFile').yields(null, 'name: 1\n');

      return Descriptor.read(Descriptor, SCHEMA_PATH, FILE_PATH)
      .catch((errors) => {
        assert.ok(Array.isArray(errors));
        assert.strictEqual(errors.length, 1);
      });
    });

    it('returns error if data is not valid yaml or json', function() {
      sandbox.stub(fs, 'readFile').yields(null, '{"a":1}');

      return Descriptor.read(Descriptor, SCHEMA_PATH, FILE_PATH)
      .catch((errors) => {
        assert.ok(Array.isArray(errors));
        assert.strictEqual(errors.length, 2);
      });
    });

    it('returns an error if it cannot read the file', function() {
      sandbox.stub(fs, 'readFile').yields(new Error('boo'));

      return Descriptor.read(Descriptor, SCHEMA_PATH, FILE_PATH)
      .catch((errors) => {
        assert.ok(Array.isArray(errors));
        assert.strictEqual(errors.length, 1);
        assert.strictEqual(errors[0].message, 'boo');
      });
    });
  });
});
