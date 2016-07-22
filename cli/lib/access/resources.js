'use strict';

// TODO: Add comments
// TODO: More descriptive var names
// TODO: Constants to object
// TODO: Have path validation except a path and/or path + secret

var _ = require('lodash');
var cpath = require('common/cpath');

var OR_EXP = require('common/cpath/definitions').OR_EXP_REGEX;
var SLUG_OR_WILDCARD = require('common/cpath/definitions').SLUG_OR_WILDCARD_REGEX;

var PROJECT = 'project';
var ENVIRONMENT = 'environment';
var SERVICE = 'service';
var IDENTITY = 'identity';
var INSTANCE = 'instance';
var SECRET = 'secret';

var RESOURCES = [PROJECT, ENVIRONMENT, SERVICE, IDENTITY, INSTANCE, SECRET];

var OR_ELIGIBLE = {};

OR_ELIGIBLE[PROJECT] = true;
OR_ELIGIBLE[ENVIRONMENT] = true;
OR_ELIGIBLE[SERVICE] = true;
OR_ELIGIBLE[INSTANCE] = true;

var resources = exports;

resources.explode = function (map) {
  return resources.expand(map, true);
};

resources.expand = function (map, explode) {
  explode = explode || false;

  var resourcePerms = [];
  var path = [];
  var segments = _.clone(RESOURCES);

  expandResource(segments, path, map); // eslint-disable-line no-use-before-define

  function expandResource(s, p, m) {
    if (s.length === 0) {
      return;
    }

    var resource = _.take(s)[0];
    var value = m[resource];

    if (OR_ELIGIBLE[resource] && OR_EXP.test(value)) {
      var values = _.chain(value)
        .trimStart('[')
        .trimEnd(']')
        .split('|')
        .value();

      _.each(values, function (v) {
        var newSeg = _.clone(s);
        var newPath = _.clone(p);
        var newMap = _.clone(m);

        newMap[resource] = v;
        expandResource(newSeg, newPath, newMap);
      });
      return;
    }

    p.push(value);

    if (explode) {
      resourcePerms.push('/' + p.join('/'));
    }

    if (!explode && p.length === 6) {
      resourcePerms.push('/' + p.join('/'));
    }

    expandResource(_.slice(s, 1), p, m);
  }

  return resourcePerms;
};

resources.validPath = function (path, secret) {
  if (!cpath.validateExp(path)) {
    return new Error('Invalid path provided');
  }

  if (!SLUG_OR_WILDCARD.test(secret)) {
    return new Error('Invalid secret provided');
  }

  return true;
};

resources.fromPath = function (path, secret) {
  var cpathObj = cpath.parseExp(path);
  var resourceMap = _.pick(cpathObj, RESOURCES);

  resourceMap.secret = secret;

  return resourceMap;
};
