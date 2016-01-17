'use strict';

var _ = require('lodash');

module.exports = function deepPick (fullPath, obj, sep) {
  sep = sep || '.'; // default to . as path separator

  function pick (path, obj) {
    var fragments = path.split(sep);
    var fragment = fragments[0];
    var remainder = fragments.slice(1).join('.');

    if (typeof obj !== 'object') {
      throw new TypeError(
        'must be an object: '+fullPath.replace(remainder, ''));
    }

    if (obj.hasOwnProperty(fragment)) {
      if (remainder.length > 0) {
        return pick(remainder, obj[fragment]); 
      }

      return obj[fragment];
    }

    return undefined;
  }

  return pick(fullPath, obj);
};
