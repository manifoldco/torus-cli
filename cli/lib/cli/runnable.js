'use strict';

var series = require('./series');
var Middleware = require('./middleware');

function Runnable() {
  this.preHooks = [];
  this.postHooks = [];
}
module.exports = Runnable;

Runnable.prototype.hook = function (type, fn) {
  if (type !== 'pre' && type !== 'post') {
    throw new TypeError('Unknown hook type: ' + type);
  }

  var hookList = (type === 'pre') ? this.preHooks : this.postHooks;
  hookList.push(new Middleware(fn));
};

Runnable.prototype.runHooks = function (type, ctx) {
  if (type !== 'pre' && type !== 'post') {
    throw new TypeError('Unknown hook type: ' + type);
  }

  var hookList = (type === 'pre') ? this.preHooks : this.postHooks;
  var middleware = hookList.map(function (mw) {
    return mw.run.bind(mw, ctx);
  });

  return series(middleware);
};
