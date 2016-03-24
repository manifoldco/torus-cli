'use strict';

function Context (program) {

  if (!program) {
    throw new TypeError('A program must be provided');
  }

  this.slug = null;
  this.cmd = null;
  this.program = program;
  this.params = [];
  this.props = {};
}

Context.prototype.prop = function (name) {
  return this.props[name];
};

module.exports = Context;
