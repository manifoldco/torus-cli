'use strict';
var _ = require('lodash');

var log = require('../util/log').get('help').log;

module.exports.command = function help (argv) {
  log('Arigato amazing CLI tool');
}
