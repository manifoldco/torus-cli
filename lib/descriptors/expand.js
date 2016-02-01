'use strict';

var fs = require('fs');
var path = require('path');

var yaml = require('yamljs');
var _ = require('lodash');

/**
 * Given a file path it loads the file and then scans for all json reference
 * objects. It then loads the json referenced file contents into the block.
 */
module.exports = function expand (filePath) {
  return new Promise((resolve, reject) => {
    
    var base = path.dirname(filePath);
    var file = path.basename(filePath);
    return load(base, { ref: '', path: file }).then((ref) => {

      var data = ref.body;
      var toLoad = explore(data, '').map(load.bind(null, base));

      return Promise.all(toLoad).then((loaded) => {
        loaded.forEach((part) => {
          _.set(data, part.ref, part.body);
        });

        resolve(data);
      });
    }).catch(reject);
  });
};

function explore (obj, ref) {
  if (!_.isPlainObject(obj) && !Array.isArray(obj)) {
    throw new TypeError('Not an array or object: '+ref);
  }

  var toExpand = [];
  var iteratee;
  var pointer;
  if (Array.isArray(obj)) {
    pointer = '[$0]';
    iteratee = Array.apply(null, { length: obj.length })
    .map(Number.call, Number);
  } else {
    iteratee = Object.keys(obj);
    pointer = '.$0';
  }

  iteratee.forEach((k) => {
    var kRef = ref + pointer.replace('$0', k);
    if (_.isPlainObject(obj[k]) || Array.isArray(obj[k])) {
      toExpand = toExpand.concat(explore(obj[k], kRef));
      return;
    }

    if (k === '$ref') {
      // XXX we're not using kRef because we want to replace k's parent with
      // the value of the ref.
      toExpand.push({
        ref: ref,
        path: obj[k]
      });
    }
  });

  return toExpand;
}

function load (base, obj) {
  return new Promise((resolve, reject) => {
    var p = path.resolve(path.join(base, obj.path));
    fs.readFile(p, { encoding: 'utf-8' }, function (err, data) {
      if (err) {
        return reject(err); 
      }

      switch (path.extname(p)) {
        case '.json':
          data = JSON.parse(data);
          break;

        case '.yaml':
        case '.yml':
          data = yaml.parse(data);
          break;

        default:
          throw new TypeError('Extension not supported: '+p);
      }
      
      resolve({
        ref: obj.ref,
        body: data
      });
    });
  });
}
