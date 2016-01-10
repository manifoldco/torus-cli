'use strict';

const kbpgp = require('kbpgp');
const wrap = require('cbwrap').wrap;

const users = require('./users'); 
const errors = require('./errors');
const keysApi = require('./api/keys');
const triplesec = require('./util/triplesec');

const keys = exports;

const generateKeyPair = wrap(kbpgp.KeyManager.generate_rsa);

keys.create = function (params) {
  return new Promise((resolve, reject) => {

    if (!params.uuid || !params.session_token) {
      return reject(new errors.NotAuthenticatedError());
    }

    if (!params.password || !params.email) {
      return reject(new Error('Must provide proper params'));
    }

    var sudoParams = {
      email: params.email,
      password: params.password
    };

    users.sudo(sudoParams).then((credentials) => {

      const options = {
        userid: params.uuid
      };

      return generateKeyPair(options).then((keyManager) => {
        const sign = wrap(keyManager.sign.bind(keyManager));
        const resolvePromise = Promise.resolve.bind(Promise, keyManager);

        return sign({}).then(resolvePromise);
      }).then((keyManager) => {

        const exportPublic = wrap(
          keyManager.export_pgp_public.bind(keyManager));
        const exportPrivate = wrap(
          keyManager.export_pgp_private.bind(keyManager));

          return Promise.all([
            exportPublic({}),
            exportPrivate({ passphrase: params.password })
          ]); 
      }).then((keypair) => {
        var opts = {
          data: new triplesec.Buffer(keypair[1]),

          // XXX should this be a derived key?
          key: new triplesec.Buffer(params.password)
        };

        return triplesec.encrypt(opts).then((buf) => {
          return Promise.resolve([
            keypair,
            buf.toString('base64')
          ]);
        });
      }).then((keypair) => {
        // TODO: Priovate key needs to be encrypted w/ a derived key..
        return keysApi.upload({
          session_token: params.session_token,
          login_token: credentials.login_token,
          password: credentials.pwh,
          public_key: keypair[0],
          private_key: keypair[1]
        });    
      }).then((results) => {

        // TODO: Pass back pgp keys or at least the finger print
        return resolve(results);
      });
    }).catch(reject);
  }); 
};
