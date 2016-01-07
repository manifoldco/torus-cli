'use strict';

var errors = require('./errors');

class CommandInterface {
  execute () {
    return new Promise((resolve, reject) => {
      reject(new errors.NotImplementedError());
    });
  }

  cancel () {
    return new Promise((resolve, reject) => {
      reject(new errors.NotImplementedError());
    });
  }
}

exports.CommandInterface = CommandInterface;
