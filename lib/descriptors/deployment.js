'use strict';

const Descriptor = require('./descriptor').Descriptor;

const deployment = exports;

class Deployment extends Descriptor {
  constructor (path, contents) {
    super('deployment', path, contents);
  }
}

deployment.Deployment = Deployment;
