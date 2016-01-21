'use strict';

const Descriptor = require('./descriptor').Descriptor;

const service = exports;

class Service extends Descriptor {
  constructor (path, contents) {
    super('service', path, contents);
  }
}

service.Service = Service;
