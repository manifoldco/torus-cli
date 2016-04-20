'use strict';

var _ = require('lodash');
var Promise = require('es6-promise').Promise;
var minimist = require('minimist');

var util = require('util');

var Command = require('./command');;
var Context = require('./context');
var Runnable = require('./runnable');

function Program (name, version, templates) {
  Runnable.call(this);

  if (typeof name !== 'string') {
    throw new TypeError('A string must be provided for the name');
  }
  if (typeof version !== 'string') {
    throw new TypeError('A version must be provided');
  }

  if (!templates || !templates.program || !templates.command) {
    throw new TypeError('Templates must be defined');
  }

  this.name = name;
  this.version = version;
  this.templates = {
    program: _.template(templates.program, { imports: {
      '_': _
    }}),
    command: _.template(templates.command, { imports: {
      '_': _
    }})
  };
  this.commands = {};
  this.groups = {};
}

util.inherits(Program, Runnable);

module.exports = Program;

Program.prototype.command = function (cmd) {
  if (!(cmd instanceof Command)) {
    throw new TypeError('A Command object must be provided');
  }


  cmd.program = this;
  this.commands[cmd.slug] = cmd;

  if (!this.groups[cmd.group]) {
    this.groups[cmd.group] = {};
  }

  this.groups[cmd.group][cmd.subpath] = cmd;
  return this;
};

Program.prototype.help = function (ctx) {

  var params = ctx.params;
  if (!ctx.slug || (ctx.slug === 'help' && params.length === 0)) {
    return this._rootHelp(ctx);
  }

  if (ctx.slug === 'help' && params.length > 0) {
    ctx.slug = ctx.params.shift();
  }

  var c = this.commands[ctx.slug];
  if (!c) {
    console.log('Unknown command: '+ctx.slug);
    return this._rootHelp(ctx);
  }

  return this._cmdHelp(ctx, c);
};

Program.prototype._rootHelp = function root (ctx) {
  console.log(this.templates.program({
    program: this,
    ctx: ctx
  }));

  return Promise.resolve(false);
};

Program.prototype._cmdHelp = function cmd (ctx, cmd) {
  console.log(this.templates.command({
    program: this,
    ctx: ctx,
    cmd: cmd,
    group: (cmd.slug === cmd.group) ? this.groups[cmd.group] : {}
  }));

  return Promise.resolve(false);
};

Program.prototype._getSlug = function (ctx, argv) {
  argv = argv.slice(2);
  if (argv.length === 0) {
    return [];
  }

  ctx.slug = argv.shift();
  if (argv.length === 0) {
    return [];
  }

  return argv;
};

Program.prototype._parseArgs = function (ctx, cmd, argv) {
  argv = argv || [];

  var args = minimist(argv);
  ctx.params.push.apply(ctx.params, args._);
  cmd.options.forEach(function (o) {
    o.evaluate(ctx, args);
  });

  return ctx;
};

Program.prototype.run = function (argv) {

  if (!Array.isArray(argv)) {
    throw new Error('Must provide an array for argv');
  }

  var ctx = new Context(this);
  argv = this._getSlug(ctx, argv);

  if (!ctx.slug || ctx.slug === 'help') {
    ctx.params = argv;
    return this.help(ctx);
  }

  var cmd = this.commands[ctx.slug];
  if (!this.commands[ctx.slug]) {
    console.log('Unknown Command: '+ctx.slug);
    ctx.slug = ctx.group = null;
    return this.help(ctx);
  }

  ctx.cmd = cmd;
  this._parseArgs(ctx, cmd, argv);

  var self = this;
  return self.runHooks('pre', ctx).then(function() {
    return cmd.run.call(cmd, ctx).then(function() {
      return self.runHooks('post', ctx);
    });
  });
};
