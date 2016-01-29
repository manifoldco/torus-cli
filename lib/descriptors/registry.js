'use strict';

const fs = require('fs');
const path = require('path');

const Deployment = require('./deployment').Deployment;

const registry = exports;

const DEPLOYMENTS_DIR = path.resolve(path.join(__dirname, '../../deployments'));

class Registry {
  constructor () {
    this._services = null;
  }

  services () {
    return new Promise((resolve, reject) => {

      if (this._services) {
        return resolve(this._services);
      }

      // Returns a list of deployment directories
      fs.readdir(DEPLOYMENTS_DIR, function (err, folders) {
        if (err) {
          return reject(err);
        }

        var getFiles = folders.map((folder) => {
          // Right now we only support one deployment per service
          return Deployment.read(path.join(
            DEPLOYMENTS_DIR, folder, 'default.yml'));
        });

        Promise.all(getFiles).then((services) => {
          var map = {};
          services.forEach((deployment) => {
            map[deployment.get('name')] = deployment;
          });

          this._services = map;
          resolve(map);
        }).catch(reject);
      }.bind(this));  
    });
  }


  get (name) {
    return new Promise((resolve, reject) => {
      if (this._services) {
        return resolve(this._services[name]);
      }

      return this.services().then((list) => {
        resolve(list[name]); 
      }).catch(reject);
    }); 
  }
}

registry.Registry = Registry;
