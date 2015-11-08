'use strict';
var fs = require('fs');

var cmds = {};

fs.readdirSync(__dirname).forEach(function (name) {

  name = name.replace(/\..*$/,'');
  if (name === 'index') {
    return;
  }

  cmds[name] = require('./' + name);
});

module.exports = cmds;
