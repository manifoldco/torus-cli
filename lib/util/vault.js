'use strict';

const path = require('path');
const fs = require('fs');

const cbwrap = require('cbwrap');
const netrc = require('netrc');

const vault = exports;

const readFile = cbwrap.wrap(fs.readFile);
const writeFile = cbwrap.wrap(fs.writeFile);

class Vault {
  constructor(value, domain, vaultPath) {
    if (typeof value !== 'object') {
      throw new TypeError('Must provide an object');
    }

    if (typeof domain !== 'string') {
      throw new TypeError('Must provide a valid domain');
    }

    this.vault = value;
    this.domain = domain;
    this.vaultPath = vaultPath;
  }

  get (key) {
    if (!this.vault[this.domain]) {
      return undefined;
    }

    return this.vault[this.domain][key];
  }

  set (key, value) {

    if (!this.vault[this.domain]) {
      this.vault[this.domain] = {};
    }

    this.vault[this.domain][key] = value;
  }

  remove (key) {
    if (!this.vault[this.domain][key]) {
      throw new ReferenceError(`Key ${key} does not exist in vault`);
    }

    delete this.vault[this.domain][key];
  }

  all () {
    return this.vault[this.domain];
  }

  save () {
    var str = netrc.format(this.vault);
    return writeFile(this.vaultPath, str);
  }

  static readFrom (vaultPath, domain) {
    return new Promise((resolve, reject) => {

      vaultPath = path.resolve(__dirname, vaultPath);
      return readFile(vaultPath, { encoding: 'utf-8' }).then((data) => {
        data = netrc.parse(data);
        resolve(new Vault(data, domain, vaultPath));
      }).catch((err) => {
        if (err && err.code === 'ENOENT') {
          return resolve(new Vault({}, domain, vaultPath));
        }

        reject(err);
      });
    });
  }
}

vault.Vault = Vault;

var waitingPromise;
vault.get = function () {
  if (waitingPromise) {
    return waitingPromise;
  }

  waitingPromise = new Promise((resolve, reject) => {
    var vaultPath = process.env.HOME || process.env.HOMEPATH ||
                    process.env.USERPROFILE;

    vaultPath = path.join(vaultPath, '.netrc');
    Vault.readFrom(vaultPath, 'arigato.sh').then((vault) => {
      return resolve(vault);
    }).catch(reject);
  });

  return waitingPromise;
};
