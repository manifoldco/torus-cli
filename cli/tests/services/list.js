/* eslint-env mocha */

'use strict';

var assert = require('assert');

var sinon = require('sinon');
var Promise = require('es6-promise').Promise;
var utils = require('common/utils');

var Session = require('../../lib/session');
var Context = require('../../lib/cli/context');

var client = require('../../lib/api/client').create();
var serviceList = require('../../lib/services/list');

var ORG = {
  id: utils.id('org'),
  body: {
    name: 'my-org'
  }
};

var PROJECTS = [
  {
    id: utils.id('project'),
    body: {
      name: 'api-1',
      org_id: ORG.id
    }
  },
  {
    id: utils.id('project'),
    body: {
      name: 'api-2',
      org_id: ORG.id
    }
  }
];

var SERVICES = [
  {
    id: utils.id('service'),
    body: {
      name: 'api-1',
      org_id: ORG.id,
      project_id: PROJECTS[0].id
    }
  },
  {
    id: utils.id('service'),
    body: {
      name: 'api-2',
      org_id: ORG.id,
      project_id: PROJECTS[1].id
    }
  },
  {
    id: utils.id('service'),
    body: {
      name: 'api-2-1',
      org_id: ORG.id,
      project_id: PROJECTS[1].id
    }
  }
];


describe('Services List', function () {
  var CTX;
  before(function () {
    this.sandbox = sinon.sandbox.create();
  });
  beforeEach(function () {
    this.sandbox.stub(client, 'get')
      .onFirstCall()
      .returns(Promise.resolve({
        body: [ORG]
      }));
    CTX = new Context({});
    CTX.session = new Session({
      token: 'abcdefgh',
      passphrase: 'hohohoho'
    });
    CTX.params = ['hi-there'];
    CTX.options = {
      org: {
        value: ORG.body.name
      },
      project: {
        value: undefined
      }
    };
  });

  afterEach(function () {
    this.sandbox.restore();
  });

  describe('#execute', function () {
    it('requires an auth token', function () {
      CTX.session = null;

      return serviceList.execute(CTX).then(function () {
        assert.ok(false, 'should not resolve');
      }).catch(function (err) {
        assert.ok(err);
        assert.strictEqual(err.message, 'Session object not on Context');
      });
    });

    it('does not throw if the user has no services or projects', function () {
      client.get.returns(Promise.resolve({ body: [] }));
      return serviceList.execute(CTX).then(function (payload) {
        assert.deepEqual(payload, {
          projects: [],
          services: []
        });
      });
    });

    it('handles not found from api', function () {
      var err = new Error('service not found');
      err.type = 'not_found';
      client.get.returns(Promise.reject(err));

      return serviceList.execute(CTX).then(function () {
        assert.ok(false, 'should not resolve');
      }).catch(function (e) {
        assert.strictEqual(e.type, 'not_found');
      });
    });

    it('resolves with services and projects', function () {
      client.get.onCall(1).returns(Promise.resolve({ body: PROJECTS }));
      client.get.onCall(2).returns(Promise.resolve({ body: SERVICES }));

      return serviceList.execute(CTX).then(function (payload) {
        assert.deepEqual(payload, {
          projects: PROJECTS,
          services: SERVICES
        });
      });
    });

    it('returns an error if project is unknown', function () {
      client.get.returns(Promise.resolve({ body: [] }));

      CTX.options.project.value = 'api-3';

      return serviceList.execute(CTX).then(function () {
        assert.ok(false, 'should not resolve');
      }).catch(function (err) {
        assert.ok(err instanceof Error);
        assert.strictEqual(err.message, 'project not found: api-3');
      });
    });

    it('resolves with services and proejcts w/ name provided', function () {
      client.get.onCall(1).returns(Promise.resolve({ body: PROJECTS }));
      client.get.onCall(2).returns(Promise.resolve({ body: SERVICES }));

      CTX.options.project.value = PROJECTS[0].body.name;

      return serviceList.execute(CTX).then(function (payload) {
        assert.deepEqual(payload, {
          projects: [PROJECTS[0]],
          services: [SERVICES[0]]
        });
      });
    });
  });
});
