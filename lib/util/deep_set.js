'use strict';

module.exports = function deepSet (fullPath, value, obj, sep) {
  sep = sep || '.'; // default to . as path separator
  obj = obj || {};

  function set (path, obj) {
    var fragments = path.split(sep);
    var fragment = fragments[0];
    var remainder = fragments.slice(1).join('.');

    if (typeof obj !== 'object') {
      throw new TypeError(
        'must be an object: '+fullPath.replace(remainder, ''));
    }

    if (!obj.hasOwnProperty(fragment)) {
      obj[fragment] = {};
    }
    
    if (remainder.length > 0) {
      return set(remainder, obj[fragment]);
    }

    obj[fragment] = value;
  }

  set(fullPath, obj);
  return obj;
};
