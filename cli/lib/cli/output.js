'use strict';

var output = exports;

var _ = require('lodash');

/**
 * Passthrough generator
 *
 * @param {function} fn
 */
output.create = function(fn) {
  return function(noTop, noBottom) {
    if (!_.isFunction(fn)) {
      throw new Error('missing output function');
    }

    // TODO: Proper output module for errors and banner messages
    if (!noTop) { console.log(''); }
    fn();
    if (!noBottom) { console.log(''); }
  };
};
