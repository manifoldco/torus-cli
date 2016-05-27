'use strict';

var wrap = require('./wrap');

function Middleware(fn) {
  if (typeof fn !== 'function') {
    throw new TypeError('Middleware must be a function');
  }

  this.fn = fn;
}

Middleware.prototype.run = function (ctx) {
  return wrap(this.fn.bind(null, ctx));
};

module.exports = Middleware;
