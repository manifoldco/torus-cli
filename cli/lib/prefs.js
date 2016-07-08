'use strict';

var _ = require('lodash');
var path = require('path');

// [default] in the .arigatorc
var DEFAULT_SECTION = 'default';

// Map of valid sections, properties and validation functions
var SECTIONS = {
  default: {
    environment: _.isString,
    service: _.isString
  },
  core: {
    context: _.isBoolean
  }
};

/**
 * Return true if the section and key are valid, and the validation function returns true
 * @returns {Bool}
 */
function valid(section, key, value) {
  return SECTIONS[section] && SECTIONS[section][key] && SECTIONS[section][key](value);
}

function defineProperty(obj, key) {
  Object.defineProperty(obj, key, {
    get: function () {
      return obj._values[key];
    },
    configurable: false,
    enumerable: true
  });
}

/**
 * Validate and manage the contents of the .arigatorc file
 * @param {String} absoute path to the .arigatorc file
 * @param {Object} contents of the .arigatorc file
 */
function Prefs(rcPath, rcContents) {
  if (!_.isString(rcPath)) {
    throw new TypeError('Must provide rcPath string');
  }

  if (!path.isAbsolute(rcPath)) {
    throw new Error('Must provide an absolute rc path');
  }

  if (!_.isPlainObject(rcContents)) {
    throw new TypeError('Must provide an rc object');
  }

  this.path = rcPath;
  this._rc = rcContents;
  this._values = {};

  var self = this;
  _.keys(SECTIONS).forEach(function (section) {
    var preferences = rcContents[section] || {};

    var validPreferences = _.every(preferences, function (prefVal, prefKey) {
      return valid(section, prefKey, prefVal);
    });

    if (!validPreferences) {
      throw new Error('The contents of the .arigatorc (' + rcPath + ') are not valid');
    }

    if (_.keys(preferences).length < 1) return;

    self._values[section] = preferences;
    defineProperty(self, section);
  });

  Object.defineProperty(self, 'values', {
    get: function () {
      return self._values;
    },
    configurable: false,
    enumerable: true
  });
}

Prefs.SECTIONS = SECTIONS;
Prefs.DEFAULT_SECTION = DEFAULT_SECTION;

/**
 * Convert and validate dot seperated preference eg. default.context and set preference
 * Section defaults to [default]
 * @param {String} preference or dot seperated [section|preference].
 * @param {Object} new preference value
 */
Prefs.prototype.set = function (sectionKey, value) {
  var segments = sectionKey.split('.');

  // Accept only [section][key]
  if (segments.length >= 3) {
    throw new Error('Invalid sectionKey provided');
  }

  var section = segments[0];
  var key = segments[1];

  // [default] to default
  if (!key) {
    key = section;
    section = DEFAULT_SECTION;
  }

  if (!SECTIONS[section]) {
    throw new Error('Invalid section of preferences provided.');
  }

  if (!SECTIONS[section][key]) {
    throw new Error('Invalid preference provided.');
  }

  var validPreference = valid(section, key, value);

  if (!validPreference) {
    throw new Error('Invalid value for preference provided.');
  }

  this._values[section] = this._values[section] || {};
  this._values[section][key] = value;
};

module.exports = Prefs;
