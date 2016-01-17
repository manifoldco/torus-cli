'use strict';

const clone = require('clone');
const yaml = require('yamljs');
const fs = require('fs');
const path = require('path');

const ArigatoConfigError = require('./errors').ArigatoConfigError;
const deepPick = require('./util/deep_pick');
const deepSet = require('./util/deep_set');
const resolver = require('./util/resolver');
const schema = require('./util/schema');

const arigato = exports;

const FILE_NAME = 'arigato.yml';

class Arigato {
  constructor (path, contents) {

    if (typeof path !== 'string' || typeof contents !== 'object') {
      throw new TypeError('Path must be a string, contents must be an object');
    }

    this.path = path;
    this.contents = contents;
  }

  get (key) {
    var data = deepPick(key, this.contents);
    if (data !== undefined) {
      return clone(data);
    }

    var schemaContents = schema.get('arigato');
    var segment = schemaContents.properties[key];
    if (!segment) {
      throw new ReferenceError('Unknown key in schema props: '+key);
    }

    // It'd be ncie if this did deep properties; but it's unneeded for now
    switch (segment.type) {
      case 'object':
        return {};
      case 'array':
        return [];
      default:
        return null;
    }
  }

  set (key, value) {
    return new Promise((resolve, reject) => {
      // We don't want to modift `this.contents` until we've verified whether or
      // not the contents meet our schema.
      var contents = deepSet(key, value, clone(this.contents));
      Arigato.validate(contents).then((contents) => {
        this.contents = contents;
        resolve(this.contents);
      }).catch(reject);
    });
  }

  write () {
    return new Promise((resolve, reject) => {
      // Check that we can find, read and write to the file
      var filePath = this.path;

      var contents = this.yaml;
      fs.writeFile(filePath, contents, function(err) {
        if (err) {
          return reject(new ArigatoConfigError(
            'Could not write '+FILE_NAME+' file: '+filePath
          ));
        }

        resolve();
      });
    });
  }

  get yaml () {
    // Inline after a depth of 5 and use two sapces for indentation.
    return yaml.stringify(this.contents, 5, 2);
  }

  static validate (input) {
    return new Promise((resolve, reject) => {
      input = input || '{}';
      return schema.validate('arigato', input).then(resolve).catch((errors) => {
        if (!Array.isArray(errors) && errors.name !== 'ValidationError') {
          return reject(errors);
        }

        errors = errors.map((err) => {
          return new ArigatoConfigError(err.message, err.code);
        });

        reject(errors);
      });
    });
  }

  static create (base, opts) {
    return new Promise((resolve, reject) => {

      var filePath = path.join(base, FILE_NAME);
      var ag = new Arigato(filePath, {
        owner: opts.user_id,
        app: opts.app.uuid
      });

      ag.write().then(resolve).catch(reject);
    });
  }

  static find (base) {
    return new Promise((resolve, reject) => {
      resolver.parents(base, FILE_NAME).then((files) => {
        // The first one (so index 0) is the closest and therefore most relevant
        if (files.length === 0) {
          return resolve(null); // we didn't find anything
        }

        var file = files[0];
        fs.readFile(file, { encoding: 'utf-8' }, function (err, data) {
          if (err) {
            return reject(err);
          }

          try {
            data = yaml.parse(data);
          } catch (err) {
            return reject(err);
          }

          return Arigato.validate(data).then((data) => {
            resolve(new Arigato(file, data));
          }).catch(reject);
        });
      }).catch(reject);
    });
  }
}

arigato.Arigato = Arigato;
arigato.FILE_NAME = FILE_NAME;
