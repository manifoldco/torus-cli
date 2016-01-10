'use strict';

var wrap = require('cbwrap').wrap;
var ts = require('triplesec');

const triplesec = exports;

triplesec.Buffer = ts.Buffer;
triplesec.encrypt = wrap(ts.encrypt);
triplesec.decrypt = wrap(ts.decrypt);
