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
    return this.vault[this.domain][key];
  }

  set (key, value) {

    if (!this.vault[this.domain]) {
      this.vault[this.domain] = {};
    }

    this.vault[this.domain][key] = value;
  }

  all () {
    return this.vault[this.domain];
  }

  save () {
    return new Promise(() => {
      var str = netrc.format(this.vault);
      return writeFile(this.vaultPath, str);
    });
  }

  static readFrom (vaultPath, domain) {
    return new Promise(() => {
      vaultPath = path.resolve(__dirname, vaultPath);
      return readFile(vaultPath, { encoding: 'utf-8' }).then((data) => {
        data = netrc.parse(data);
        return Promise.resolve(new Vault(data, domain, vaultPath)); 
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

  waitingPromise = new Promise(() => {
    var vaultPath = process.env.HOME || process.env.HOMEPATH || 
                    process.env.USERPROFILE;
    
    vaultPath = path.join(vaultPath, '.netrc');
    return Vault.readFrom(vaultPath, 'arigato.sh');
  });

  return waitingPromise;
};
