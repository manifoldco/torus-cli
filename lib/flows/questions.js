'use strict';

const _ = require('lodash');
const inquirer = require('inquirer');

const promiseUtil = require('../util/promise');
const log = require('../util/log').get('questions');

const questions = exports;

/**
 * Derives the qustions that must be asked of the users and then prompts them
 * for that information.
 */
questions.prompt = function (descriptor) {
  var inquiries = questions.derive(descriptor);
  inquiries = _.mapValues(inquiries, (inquiry) => {
    return new Promise((resolve) => {
      log.print(inquiry.description+':\n');
      inquirer.prompt(inquiry.questions, function (answers) {
        resolve({
          type: inquiry.type,
          name: inquiry.name,
          answers: answers 
        });
      });
    });
  });

  // inquiries is an object now so we need to run all the promises
  // that resolve the tree..
  return promiseUtil.map(inquiries);
};

/**
 * Converts output from inquirier.prompt into the cred tree format (which has
 * not yet been defined)
 */
questions.toCredTree = function (descriptor, inquiries) {
  var credTree = {};
  var serviceName = descriptor.get('name');
  credTree[serviceName] = {};

  Object.keys(inquiries).forEach((name) => {
    var inquiry = inquiries[name];
    if (!credTree[serviceName][name]) {
      credTree[serviceName][name] = {};
    }

    Object.keys(inquiry.answers).forEach((key) => {
      credTree[serviceName][inquiry.name][key] = inquiry.answers[key]; 
    });
  });

  return credTree;
};

/**
 * Derives the questions that must be asked from the user for us to support the
 * given descriptor. Outputs a format used by the inquirer module from node.
 */
questions.derive = function (descriptor) {

  var credentials = descriptor.get('credentials');
  var provided = descriptor.get('default.provided');
  var inquiries = {};

  Object.keys(provided).forEach((name) => {
    var requirement = provided[name];
    var inquiry = {
      name: name,
      description: requirement.description,
      type: requirement.type,
      questions: []
    };

    var credential = credentials.find((cred) => {
      return cred.name === requirement.type;
    });

    if (!credential) {
      // TODO: This shouldn't happen if the schema is valid figure out
      // how to make the schema handle this for us.
      throw new Error('Could not find credential: '+requirement.type);
    }
 
    // TODO: Support validation by running it through a json schema validator
    var properties = credential.schema.properties;
    _.each(properties, (value, key) => {
      inquiry.questions.push({
        type: value['x-input-type'],
        name: key,
        message: value.description
      }); 
    });

    inquiries[name] = inquiry;
  });

  return inquiries;
};
