'use strict';

const fs = require('fs');

const _ = require('lodash');
const clone = require('clone');
const yaml = require('yamljs');

const schema = require('../util/schema');

const descriptor = exports;

class Descriptor {
  constructor (schema, path, contents) {
    if (typeof schema !== 'string' || typeof path !== 'string' ||
        typeof contents !== 'object') {
      throw new TypeError('Path must be a string, contents must be an object');
    }

    this.schema = schema;
    this.path = path;
    this.contents = contents;
  }

  get (key) {
    var data = _.get(this.contents, key, undefined);
    if (data !== undefined) {
      return clone(data);
    }

    var schemaContents = schema.get(this.schema);
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

  /**
   * Supports either set(key, value) where key is a path or a map like:
   *  
   *  {
   *    key: value,
   *    key: value
   *  }
   */
  set (key, value) {
    return new Promise((resolve, reject) => {
      var data = {};
      if (typeof key === 'string') {
        data[key] = value;
      } else {
        data = key;
      }

      // We don't want to modift `this.contents` until we've verified whether or
      // not the contents meet our schema.
      var contents = clone(this.contents);
      _.each(data, (v,k) => {
        _.set(contents, k, v);
      });

      this.validate(contents).then((contents) => {
        this.contents = contents;
        resolve(clone(this.contents));
      }).catch(reject);
    });
  }

  validate (contents) {
    contents = contents || this.contents;
    return Descriptor.validate(this.schema, contents);
  }

  get yaml () {
    // Inline after a depth of 5 and use two sapces for indentation.
    return yaml.stringify(this.contents, 5, 2);
  }

  write () {
    return new Promise((resolve, reject) => {
      var filePath = this.path;
      var contents = this.yaml;

      fs.writeFile(filePath, contents, function (err) {
        if (err) {
          return reject(err);
        }

        resolve();
      });
    }); 
  }

  static create (Klass, schema, filePath, data) {
    var obj = new Klass(schema, filePath, data);
    return obj.write();
  }

  static validate (schemaName, data) {
    return schema.validate(schemaName, data);
  }

  static read (Klass, schemaName, filePath) {
    return new Promise((resolve, reject) => {
      fs.readFile(filePath, { encoding: 'utf-8' }, (err, data) => {
        if (err) {
          return reject([err]);
        }

        try {
          data = yaml.parse(data);
        } catch (err) {
          return reject([err]);
        }

        return schema.validate(schemaName, data).then((data) => {
          if (Klass === Descriptor) {
            return resolve(new Klass(schemaName, filePath, data));
          }

          resolve(new Klass(filePath, data));
        }).catch(reject);
      });
    });
  }
}

descriptor.Descriptor = Descriptor;
