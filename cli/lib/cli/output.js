'use strict';

var output = exports;

var _ = require('lodash');

/**
 * Passthrough generator
 *
 * @param {function} fn
 */
output.create = function (fn) {
  if (!_.isFunction(fn)) {
    throw new Error('missing output function');
  }

  return function (opts) {
    opts = opts || {};

    // TODO: Proper output module for errors and banner messages
    if (_.isUndefined(opts.top) || opts.top === true) { console.log(''); }

    // Pass all args after padding preferences to the output function
    var args = Array.prototype.slice.call(arguments, 1);
    fn.apply(null, args);

    if (_.isUndefined(opts.bottom) || opts.bottom === true) { console.log(''); }
  };
};

/**
 * Handle output of error objects
 *
 * @param {object} err
 */
output.error = function (err) {
  var prefix = 'Fatal error';
  var type = err && err.type ? err.type : 'unknown';

  if (type === 'validation') {
    prefix = 'Validation error';
  } else if (type === 'unauthorized') {
    prefix = 'Authentication error';
  } else if (type === 'usage') {
    prefix = 'Invalid usage';
  }

  if (process.env.NODE_ENV === 'arigato' && err.stack) {
    console.error(err.stack);
  } else {
    console.error(prefix + ':\n\t', err.message || err, '\n');
    if (type === 'unknown' && err.stack) {
      console.error(err.stack);
    }
  }
};
