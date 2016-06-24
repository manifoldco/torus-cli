'use strict';

var Promise = require('es6-promise').Promise;

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


  function runHooks(list) {
    var item = list.shift();
    if (!item) {
      return Promise.resolve(0);
    }

    return item().then(function (exitCode) {
      if (exitCode !== 0) {
        return exitCode;
      }

      return runHooks(list);
    });
  }

  return runHooks(middleware);
};
