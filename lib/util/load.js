'use strict';

var fs = require('fs');
var path = require('path');

var yaml = require('yamljs');
const load = exports;

/**
 * Given a file path loads it and then attempts to parse the contents based
 * on the file extension.
 */
load.file = function file (filePath) {
  return new Promise((resolve, reject) => {
    fs.readFile(filePath, { encoding: 'utf-8' }, function (err, data) {
      if (err) {
        return reject(err); 
      }

      try {
        switch (path.extname(filePath)) {
          case '.json':
            data = JSON.parse(data);
          break;

          case '.yaml':
          case '.yml':
            data = yaml.parse(data);
          break;

          default:
            throw new TypeError('Extension not supported: '+filePath);
        }
      } catch (err) {
        return reject(err);
      }
      
      resolve(data);
    });
  });
};
