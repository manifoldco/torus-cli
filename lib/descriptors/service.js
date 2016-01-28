'use strict';

const path = require('path');
const wrap = require('cbwrap').wrap;
const readFile = wrap(require('fs').readFile);

const Descriptor = require('./descriptor').Descriptor;

const service = exports;

class Service extends Descriptor {
  constructor (path, contents) {
    super('service', path, contents);
  }

  initialize () {

    var self = this;
    function resolvePath (filePath, root) {
      filePath = filePath.replace(
        '$DESCRIPTOR_HOME', path.dirname(root));
      filePath = filePath.replace('file://', '');
      return path.resolve(filePath);
    }

    function expand (cred) {
      return new Promise((resolve, reject) => {
        var filePath = resolvePath(cred.schema, self.path);

        return readFile(filePath, { encoding: 'utf-8'}).then((data) => {
          try {
            data = JSON.parse(data);
          } catch (err) {
            return reject(err);
          }
        
          resolve(data);
        }).catch(reject);
      });
    }

    var credentials = this.get('credentials');
    var fetchCredentials = credentials.map((cred, i) => {
      return new Promise((resolve, reject) => {
        expand(cred).then((data) => {

          var update = {};
          update[`credentials[${i}].schema`] = data;

          if (!cred.methods) {
            return self.set(update);
          }

          cred.methods.forEach((method, k) => {
            update[`credentials[${i}].methods[${k}].src`] = resolvePath(
              method.src, self.path);
          });

          return self.set(update);
        }).then(resolve).catch(reject);
      });
    });

    return Promise.all(fetchCredentials);
  }

  static read (filePath) {
    return new Promise((resolve, reject) => {
      return super.read(Service, 'service', filePath).then((descriptor) => {
        return descriptor.initialize().then(resolve.bind(null, descriptor)); 
      }).catch(reject);
    });
  }
}

service.Service = Service;
