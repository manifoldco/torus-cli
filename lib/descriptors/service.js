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
    function expand (cred) {
      return new Promise((resolve, reject) => {
        var filePath = cred.schema;
        filePath = filePath.replace(
          '$DESCRIPTOR_HOME', path.dirname(this.path));
        filePath = filePath.replace('file://', '');

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
        expand.call(this, cred).then((data) => {
          return this.set(`credentials[${i}].schema`, data);
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
