'use strict';

var assert = require('assert');

var schema = require('../../lib/util/schema');

var schemaPath = '../test/data/schema';

function cannotResolve () {
  assert.ok(false, 'Must not resolve');
}

describe('schema', function() {
  describe('#validate', function() {
    it('returns an error for invalid data', function () {
      var input = {
        name: null
      };

      schema.validate(schemaPath, input)
        .then(cannotResolve).catch(function(err) {
        assert.ok(err);
      });
    });

    it('returns an error for missing data', function() {
      var input = {};

      schema.validate(schemaPath, input)
        .then(cannotResolve).catch(function(err) {
        assert.ok(err);
      });
    });

    it('returns an error if there is an additional property', function() {
      var input = { b: false };

      schema.validate(schemaPath, input)
        .then(cannotResolve).catch(function(err) {
        assert.ok(err);
      });
    });

    it('passes if data is valid', function() {
      var input = { name: 'abcd' };

      return schema.validate(schemaPath, input).then(function(data) {
        assert.strictEqual(data.name, 'abcd');
      });
    });

    it('allows you to pass in object as schema', function () {
      var mySchema = {
        '$schema': 'https://json-schema.org/draft-04/schema#',
        title: 'my test schema',
        type: 'object',
        properties: {
          a: { type: 'string' }
        },
        required: [ 'a' ]
      };

      return schema.validate(mySchema, { a: 'b'}, (data) => {
        assert.strictEqual(data.a, 'b');
      });
    });
  });

  describe('formats', function() {

    it('supports slug format', function() {
      var input = { name: 'ab' };

      return schema.validate(schemaPath, input)
        .then(cannotResolve).catch(function(errors) {

        assert.ok(errors);
        assert.strictEqual(errors.length, 1);
        assert.ok(/Invalid slug/.test(errors[0].message));
      });
    });
  });
});
