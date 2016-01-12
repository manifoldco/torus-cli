'use strict';

const servicesApi = require('./api/services');

const services = exports;

services.supported = function () {
  return servicesApi.supported();
};
