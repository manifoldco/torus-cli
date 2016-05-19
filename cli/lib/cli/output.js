'use strict';

var output = exports;

var _ = require('lodash');

/**
 * Passthrough generator
 *
 * @param {function} fn
 */
output.create = function(fn) {
  if (!_.isFunction(fn)) {
    throw new Error('missing output function');
  }

  return function(opts) {
    opts = opts || {};

    // TODO: Proper output module for errors and banner messages
    if (_.isUndefined(opts.top) || opts.top === true) { console.log(''); }

    // Pass all args after padding preferences to the output function
    var args = Array.prototype.slice(arguments, 1);
    fn.call(null, args);

    if (_.isUndefined(opts.bottom) || opts.bottom === true) { console.log(''); }
  };
};
