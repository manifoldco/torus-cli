'use strict';

const _ = require('lodash');

/**
 * Uses DFS to deep explore an object/array and find all json reference objects.
 *
 * These objects are then returned as a list with theri location in the object
 * and the path they are referencing.
 */
module.exports = function explore (obj, ref) {
  ref = ref || '';
  if (!_.isPlainObject(obj) && !Array.isArray(obj)) {
    throw new TypeError('Not an array or object: '+ref);
  }

  var toExpand = [];
  var iteratee;
  var pointer;
  if (Array.isArray(obj)) {
    pointer = '[$0]';
    iteratee = Array.apply(null, { length: obj.length })
    .map(Number.call, Number);
  } else {
    iteratee = Object.keys(obj);
    pointer = '.$0';
  }

  iteratee.forEach((k) => {
    var kRef = ref + pointer.replace('$0', k);
    if (_.isPlainObject(obj[k]) || Array.isArray(obj[k])) {
      toExpand = toExpand.concat(explore(obj[k], kRef));
      return;
    }

    if (k === '$ref') {
      // XXX we're not using kRef because we want to replace k's parent with
      // the value of the ref.
      toExpand.push({
        objPath: ref,
        filePath: obj[k]
      });
    }
  });

  return toExpand;
};
