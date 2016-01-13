'use strict';

const request = require('request');

const sendgrid = exports;

sendgrid.credential = function (params) {
  return new Promise((resolve, reject) => {
    if (!params.username || !params.password) {
      return reject(new Error('Missing parameters'));
    }

    var opts = {
      url: 'https://api.sendgrid.com/v3/api_keys',
      auth: {
        username: params.username,
        password: params.password,
        sendImmediately: true
      },
      json: {
        name: params.name,
        scopes: [
          "mail.send"
        ]
      }
    };

    request.post(opts, function (err, res, body) {
      if (err) {
        return reject(err);
      }

      if (res.statusCode !== 201) {
        return reject(new Error('Non 201 status code: '+res.statusCode));
      }

      resolve(body);
    });
  });
};
