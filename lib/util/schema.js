'use strict';

var path = require('path');
var tv4 = require('tv4');
var _ = require('lodash');

var load = require('./load');
var explore = require('./explore');
var validation = require('./validation');

const SCHEMA_BASE = path.resolve(__dirname, '../../schema');

const schema = exports;

var validator = tv4.freshApi();
validator.addFormat({
  slug: function (input) {
    if (validation.slug(input)) {
      return null;
    }

    return 'Invalid slug';
  }
});

schema.validator = validator;

/**
 * Loads the given schema and any sub schemas it relies upon.
 */
schema.get = function (schemaPath) {

  // TODO: Come back and memozie calls to load schema so we don't try to
  // load the same schemas over and over again inside the same tree
  var fullPath = path.join(SCHEMA_BASE, schemaPath+'.json');
  function loadSchema (fullPath) {
    return new Promise((resolve, reject) => {

      var contents = schema.validator.getSchema(fullPath);
      if (contents) {
        return resolve(contents);
      }

      var baseDir = path.dirname(fullPath);
      return load.file(fullPath).then((data) => {

        schema.validator.addSchema(fullPath, data);
        var schemaPaths = explore(data).map((ref) => {
          var extIndex = ref.filePath.indexOf('.json') + 5;
          var fileName = ref.filePath.slice(0, extIndex);
          var filePath = path.resolve(baseDir, fileName);

          return filePath;
        });

        // Find the unique schema paths and then remove any we've already
        // loaded.
        schemaPaths = _.uniq(schemaPaths).filter((filePath) => {
          return (schema.validator.getSchema(filePath) !== null);
        });

        var toLoadSchemas = schemaPaths.map(loadSchema);
        return Promise.all(toLoadSchemas).then(() => {
          resolve(data);
        });
      }).catch(reject);
    });
  }

  return loadSchema(fullPath);
};

/**
 * Validates the given properties against the provided schema. If the provided
 * schema is not an path to a schema on disk (object) then it must have all
 * references fully loaded. When a path is provided, all sub schemas are loaded.
 */
schema.validate = function (schemaPath, props) {
  return new Promise((resolve, reject) => {
    var getContents = (typeof schemaPath === 'object') ?
      Promise.resolve(schemaPath) : schema.get(schemaPath);

    getContents.then((contents) => {

      var result = schema.validator.validateMultiple(
        props, contents, true, true);

      if (result.missing.length > 0) {
        var msg =
          'Couldn\'t validate; schema(s) missing: '+result.missing.join(',');
        return reject([new Error(msg)]);
      }

      if (result.valid) {
        return resolve(props);
      }

      return reject(result.errors);
    }).catch((err) => {
      reject(Array.isArray(err) ? err : [err]);
    });
  });
};
