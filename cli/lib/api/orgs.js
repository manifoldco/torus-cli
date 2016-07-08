'use strict';

var orgs = exports;

orgs.create = function (client, data) {
  var obj = {
    body: {
      name: data.name
    }
  };

  return client.post({
    url: '/orgs',
    json: obj
  }).then(function (results) {
    return results.body;
  });
};

orgs.get = function (client, query) {
  return client.get({
    url: '/orgs',
    qs: query || {}
  }).then(function (res) {
    if (!Array.isArray(res.body)) {
      throw new Error('Expected array from API');
    }

    return res.body;
  });
};
