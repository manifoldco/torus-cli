'use strict';

const salt = exports;

const cbwrap = require('cbwrap');
const randomBytes = cbwrap.wrap(require('crypto').randomBytes);

const LENGTH = salt.LENGTH = 16;

salt.generate = function () {
  return new Promise((resolve, reject) => {
    randomBytes(LENGTH).then((buf) => {
      resolve(buf); 
    }).catch(reject); 
  });
};
