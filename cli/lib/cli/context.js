'use strict';

function Context(program) {
  if (!program) {
    throw new TypeError('A program must be provided');
  }

  this.slug = null;
  this.cmd = null;
  this.program = program;
  this.params = [];
  this.options = {};
}

Context.prototype.option = function (name) {
  return this.options[name];
};

module.exports = Context;
