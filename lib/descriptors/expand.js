'use strict';

var path = require('path');

var _ = require('lodash');
var load = require('../util/load');
var explore = require('../util/explore');

/**
 * Given a file path it loads the file and then scans for all json reference
 * objects. It then loads the json referenced file contents into the block.
 */
module.exports = function expand (filePath) {

  var basePath = path.dirname(filePath);

  // filePath is relative to the basePath
  function processFile (basePath, filePath) {
    return new Promise((resolve, reject) => {
      load.file(path.resolve(basePath, filePath)).then((data) => {
        var refsToLoad = explore(data).map(
          processRef.bind(null, basePath, data));

        Promise.all(refsToLoad).then(() => {
          resolve(data);
        });
      }).catch(reject);
    });
  }

  function processRef (basePath, data, ref) {
    return new Promise((resolve, reject) => {
      return processFile(basePath, ref.filePath).then((loaded) => {
        _.set(data, ref.objPath, loaded);
        resolve();
      }).catch(reject);
    });
  }

  return processFile(basePath, filePath);
};
