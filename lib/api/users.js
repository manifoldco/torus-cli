'use strict';

const base64url = require('base64url');
const uuid = require('uuid');

const saltGenerater = require('../util/salt').generate;
const kdfGenerator = require('../util/kdf').generate;
const errors = require('../errors');

const users = exports;

users.create = function (opts) {
  return new Promise((resolve, reject) => {

    if (!opts || !opts.email || !opts.name || !opts.password) {
      return reject(new errors.RegistryError());
    }

    saltGenerater().then((salt) => {
      return kdfGenerator(opts.password, salt).then((hashedPassword) => {
        return Promise.resolve({
          salt: salt,
          password: hashedPassword
        });
      });
    }).then((results) => {
      console.log(base64url(results.salt.toString('base64')).length,
                  base64url(results.password.toString('base64')).length);
      resolve({
        uuid: uuid.v4(),
        name: opts.name,
        email: opts.email,
        created_at: new Date(),
        updated_at: new Date()
      });  
    }).catch(reject);
  });
};
