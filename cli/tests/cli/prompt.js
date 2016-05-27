/* eslint-env mocha */

'use strict';

var _ = require('lodash');
var sinon = require('sinon');
var assert = require('assert');
var Promise = require('es6-promise').Promise;

var Prompt = require('../../lib/cli/prompt');

var MOCK_STAGES = function () {
  return [
    [
      {
        name: 'name',
        message: 'Full name',
        validate: function (input) {
          var error = 'Name too short';
          return input.length < 3 ? error : true;
        }
      }
    ],
    [
      {
        name: 'passphrase',
        type: 'password',
        message: 'Passphrase',
        validate: function (input) {
          var error = 'Password too short';
          return input.length < 8 ? error : true;
        }
      }
    ]
  ];
};

describe('Prompt', function () {
  var p;
  var promise;
  describe('constructor', function () {
    it('throws error if stages is not a function', function () {
      assert.throws(function () {
        p = new Prompt('nope');
      }, /stages must be a function/);
    });

    it('throws error if stages does not return array', function () {
      assert.throws(function () {
        p = new Prompt(function () { return null; });
      }, /stages must return array/);
    });

    it('sets default opts', function () {
      p = new Prompt(function () { return []; });
      assert.deepEqual(p.aggregate, {});
      assert.strictEqual(p._stageAttempts, 0);
      assert.strictEqual(p._stageFailed, false);
      assert.strictEqual(p._maxStageAttempts, 1);
    });
  });

  describe('prompt', function () {
    beforeEach(function () {
      p = new Prompt(MOCK_STAGES);
      promise = Promise.resolve();
    });

    afterEach(function () {
      if (p._inquirer.prompt.restore) {
        p._inquirer.prompt.restore();
      }
    });

    describe('#start', function () {
      it('calls #ask with all stages', function () {
        sinon.stub(p, 'ask').returns(Promise.resolve());
        return p.start().then(function () {
          assert.strictEqual(p.ask.callCount, 1);

          var stage = p.ask.firstCall.args[1];
          promise = p.ask.firstCall.args[0];

          assert.ok(promise instanceof Promise, 'should provide a promise');
          assert.deepEqual(stage, ['0', '1'], 'called with all stages');
        });
      });
    });

    describe('#ask', function () {
      it('compiles a list of questions', function () {
        sinon.spy(p, '_questions');
        sinon.stub(p._inquirer, 'prompt').returns(Promise.resolve());
        return p.ask(promise, _.keys(p.stages)).then(function () {
          assert.strictEqual(p._questions.callCount, 1);
          sinon.assert.calledWith(p._questions, ['0', '1']);
          var returning = p._questions.firstCall.returnValue;
          assert.deepEqual(returning.length, 2);
        });
      });

      it('aggregates responses', function () {
        sinon.spy(p, '_aggregate');
        var answers = { name: 'John Smith' };
        sinon.stub(p._inquirer, 'prompt').returns(Promise.resolve(answers));
        return p.ask(promise, _.keys(p.stages)).then(function () {
          assert.strictEqual(p._aggregate.callCount, 1);
          sinon.assert.calledWith(p._aggregate, answers);
          assert.deepEqual(p.aggregate, answers);
        });
      });

      it('recurses if stageFailed is not false', function () {
        var answers = {
          first: {
            name: 'John Smith',
            number: 2
          },
          second: {
            name: 'Jim Smith'
          }
        };

        /**
         * First attempt fails, second passes
         */
        sinon.stub(p, '_questions')
          .returns([]);

        sinon.stub(p, '_hasFailed')
          .onFirstCall()
          .returns('1')
          .onSecondCall()
          .returns(false);

        sinon.stub(p._inquirer, 'prompt')
          .onFirstCall()
          .returns(Promise.resolve(answers.first))
          .onSecondCall()
          .returns(Promise.resolve(answers.second));

        return p.ask(promise, ['0', '1']).then(function () {
          assert.strictEqual(p._hasFailed.callCount, 2, '_hasFailed once');
          // Twice for questions, once for retry message
          assert.strictEqual(p._questions.callCount, 3, '_questions once');
          assert.strictEqual(p._stageAttempts, 0);
          assert.deepEqual(p.aggregate, {
            name: 'Jim Smith',
            number: 2
          });
        });
      });
    });

    describe('#failed', function () {
      it('returns message if max attempts not met', function () {
        var message = 'some message';
        var result = p.failed(1, message);
        assert.equal(result, message);
      });

      it('returns true if max attempts not met', function () {
        var message = 'some message';
        p._stageAttempts = p._maxStageAttempts;
        var result = p.failed(1, message);
        assert.strictEqual(result, true);
      });
    });

    describe('#_reset', function () {
      it('sets stageAttempts to 0 and stageFailed to false', function () {
        p._stageAttempts = 3;
        p._stageFailed = '1';
        p._reset();
        assert.strictEqual(p._stageFailed, false);
        assert.strictEqual(p._stageAttempts, 0);
      });
    });

    describe('#_hasFailed', function () {
      it('returns true if _stageFailed not false', function () {
        p._stageFailed = '1';

        var failed = p._hasFailed();
        assert.strictEqual(failed, true);
      });

      it('returns true if _stageFailed not false', function () {
        p._stageFailed = false;

        var failed = p._hasFailed();
        assert.strictEqual(failed, false);
      });
    });
  });
});
