/* eslint-env mocha */

'use strict';

var assert = require('assert');

var sinon = require('sinon');
var Promise = require('es6-promise').Promise;
var utils = require('common/utils');

var Session = require('../../lib/session');
var api = require('../../lib/api');
var Context = require('../../lib/cli/context');
var Target = require('../../lib/context/target');

var serviceList = require('../../lib/services/list');

var ORG = {
  id: utils.id('org'),
  version: 1,
  body: {
    name: 'my-org'
  }
};

var PROJECTS = [
  {
    id: utils.id('project'),
    version: 1,
    body: {
      name: 'api-1',
      org_id: ORG.id
    }
  },
  {
    id: utils.id('project'),
    version: 1,
    body: {
      name: 'api-2',
      org_id: ORG.id
    }
  }
];

var SERVICES = [
  {
    id: utils.id('service'),
    version: 1,
    body: {
      name: 'api-1',
      org_id: ORG.id,
      project_id: PROJECTS[0].id
    }
  },
  {
    id: utils.id('service'),
    version: 1,
    body: {
      name: 'api-2',
      org_id: ORG.id,
      project_id: PROJECTS[1].id
    }
  },
  {
    id: utils.id('service'),
    version: 1,
    body: {
      name: 'api-2-1',
      org_id: ORG.id,
      project_id: PROJECTS[1].id
    }
  }
];


describe('Services List', function () {
  var ctx;
  before(function () {
    this.sandbox = sinon.sandbox.create();
  });
  beforeEach(function () {
    ctx = new Context({});
    ctx.session = new Session({
      token: 'abcdefgh',
      passphrase: 'hohohoho'
    });
    ctx.api = api.build({ auth_token: ctx.session.token });
    ctx.params = ['hi-there'];
    ctx.options = {
      org: {
        value: ORG.body.name
      },
      project: {
        value: undefined
      }
    };
    ctx.target = new Target({
      path: process.cwd(),
      context: {
        org: ORG.body.name
      }
    });

    this.sandbox.stub(ctx.api.orgs, 'get').returns(Promise.resolve([ORG]));
    this.sandbox.stub(ctx.api.services, 'get')
      .returns(Promise.resolve(SERVICES));
    this.sandbox.stub(ctx.api.projects, 'get')
      .returns(Promise.resolve(PROJECTS));
  });

  afterEach(function () {
    this.sandbox.restore();
  });

  describe('#execute', function () {
    it('requires an auth token', function () {
      ctx.session = null;

      return serviceList.execute(ctx).then(function () {
        assert.ok(false, 'should not resolve');
      }).catch(function (err) {
        assert.ok(err);
        assert.strictEqual(err.message, 'Session object not on Context');
      });
    });

    it('does not throw if the user has no services or projects', function () {
      ctx.api.services.get.returns(Promise.resolve([]));
      ctx.api.projects.get.returns(Promise.resolve([]));
      return serviceList.execute(ctx).then(function (payload) {
        assert.deepEqual(payload, {
          projects: [],
          services: []
        });
      });
    });

    it('handles not found from api', function () {
      var err = new Error('service not found');
      err.type = 'not_found';

      ctx.api.services.get.returns(Promise.reject(err));
      return serviceList.execute(ctx).then(function () {
        assert.ok(false, 'should not resolve');
      }).catch(function (e) {
        assert.strictEqual(e.type, 'not_found');
      });
    });

    it('resolves with services and projects', function () {
      return serviceList.execute(ctx).then(function (payload) {
        assert.deepEqual(payload, {
          projects: PROJECTS,
          services: SERVICES
        });
      });
    });

    it('returns an error if project is unknown', function () {
      ctx.options.project.value = 'api-3';

      ctx.api.projects.get.returns(Promise.resolve([]));
      return serviceList.execute(ctx).then(function () {
        assert.ok(false, 'should not resolve');
      }).catch(function (err) {
        assert.ok(err instanceof Error);
        assert.strictEqual(err.message, 'project not found: api-3');
      });
    });

    it('resolves with services and proejcts w/ name provided', function () {
      ctx.options.project.value = PROJECTS[0].body.name;
      return serviceList.execute(ctx).then(function (payload) {
        assert.deepEqual(payload, {
          projects: [PROJECTS[0]],
          services: [SERVICES[0]]
        });
      });
    });
  });
});
