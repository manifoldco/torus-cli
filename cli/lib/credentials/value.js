'use strict';

var value = exports;

/**
 * CredentialValue
 *
 * Object format used for serializing and deserializing credential values.
 *
 * @constructor
 * @param {Number} version
 * @param {
 */
function CredentialValue(obj) {
  obj = obj || {};
  if (obj.version !== 1) {
    throw new TypeError('version must be 1');
  }

  if (!obj.body.type || obj.body.value === undefined) {
    throw new TypeError('type and value must be defined on the body');
  }

  if (obj.body.type === 'undefined' && obj.body.value !== '') {
    throw new TypeError('Value must be empty string if type undefined');
  }

  this.version = obj.version;
  this.body = {
    type: obj.body.type,
    value: obj.body.value
  };
}
value.CredentialValue = CredentialValue;

CredentialValue.prototype.toString = function () {
  return JSON.stringify({
    version: this.version,
    body: this.body
  });
};

CredentialValue.prototype.toJSONString = CredentialValue.prototype.toString;

value.create = function (cValue) {
  return new CredentialValue({
    version: 1,
    body: {
      type: (typeof cValue),
      value: (typeof cValue === 'undefined') ? '' : cValue
    }
  });
};

value.parse = function (str) {
  if (typeof str !== 'string') {
    throw new TypeError('Must provide a string');
  }

  var obj = JSON.parse(str);
  return new CredentialValue(obj);
};
