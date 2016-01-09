'use strict';

const uuid = require('uuid');

const users = exports;

users.create = function (params) {
  return Promise.resolve({
    uuid: uuid.v4(),
    name: params.name,
    email: params.email,
    salt: params.salt,
    created_at: new Date(),
    updated_at: new Date()
  });
};
