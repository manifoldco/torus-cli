'use strict';

var series = require('./series');
var Promise = require('es6-promise').Promise;
var Middleware = require('./middleware');

function Runnable () {
  this.middleware = [];
}
module.exports = Runnable; 

Runnable.prototype.use = function (fn) {
  this.middleware.push(new Middleware(fn));
};

Runnable.prototype.runMiddleware = function (ctx) {
  var middleware = this.middleware.map(function (mw) {
    return mw.run.bind(mw, ctx);  
  });

  return series(middleware);
};
