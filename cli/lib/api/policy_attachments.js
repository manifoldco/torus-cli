'use strict';

var policyAttachments = exports;

var utils = require('common/utils');

policyAttachments.get = function (client, query) {
  return client.get({
    url: '/policy-attachments',
    qs: query || {}
  }).then(function (res) {
    return res.body;
  });
};

policyAttachments.create = function (client, data) {
  return client.post({
    url: '/policy-attachments',
    json: {
      id: utils.id('policy_attachment'),
      version: 1 || data.version,
      body: {
        org_id: data.org_id,
        owner_id: data.owner_id,
        policy_id: data.policy_id
      }
    }
  }).then(function (res) {
    return res.body;
  });
};

policyAttachments.delete = function (client, data) {
  return client.delete({
    url: '/policy-attachments/' + data.id
  }).then(function (res) {
    return res.body;
  });
};
