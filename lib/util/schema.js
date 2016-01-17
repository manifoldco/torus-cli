'use strict';

var path = require('path');
var tv4 = require('tv4');

var validation = require('./validation');

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

schema._cache = {};

schema._getPath = function (schemaPath) {
  return path.resolve(path.join(__dirname, '../../schema', schemaPath+'.json'));
};

schema.get = function (schemaPath) {
  var fullPath = schema._getPath(schemaPath);

  if (!schema._cache[fullPath]) {
    schema._cache[fullPath] = require(fullPath);
  }

  return schema._cache[fullPath];
};

schema.validate = function (schemaPath, props) {
  return new Promise((resolve, reject) => {
    var contents = schema.get(schemaPath);

    var result = schema.validator.validateMultiple(props, contents);
    if (result.valid) {
      return resolve(props);
    }

    return reject(result.errors.concat(result.missing));
  });
};
