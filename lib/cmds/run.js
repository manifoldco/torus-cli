'use strict';

const _ = require('lodash');
const child_process = require('child_process');

const log = require('../util/log').get('cmds/run');
const CommandInterface = require('../command').CommandInterface;

const errors = require('../errors');
const users = require('../users');
const envs = require('../envs');
const Arigato = require('../arigato').Arigato;
const credentials = require('../credentials');

class Run extends CommandInterface {

  execute (argv) {
    return new Promise((resolve, reject) => {
      var cmd = argv._.join(' ');
      users.me().then((user) => {
        return Arigato.find(process.cwd()).then((arigato) => {
          if (!arigato) {
            log.error('Cannot find arigato.yml file');
            return resolve(false);
          }

          return envs.list({ app_id: arigato.app }).then((appEnvs) => {
            var usersEnv = appEnvs.filter((env) => {
              return env.name === user.uuid;
            });

            if (usersEnv.length !== 1) {
              var msg = 'No environment exists for this app and user';
              return reject(new Error(msg));
            }

            // TODO now get the creds for this environment...
            var opts = {
              type: 'env',
              owner: usersEnv[0].uuid
            };
            return credentials.list(opts).then((creds) => {

              var credsMap = {};
              creds.forEach((credential) => {
                credsMap[credential.name] = credential.value;
              });

              return this.spawn({
                cmd: cmd,
                user: user,
                env: usersEnv[0],
                credentials: credsMap
              }).then((exitCode) => {
                resolve(exitCode === 0);
              });
            }).catch(reject);
          });
        });
      }).catch((err) => {
        if (err instanceof errors.NotAuthenticatedError) {
          log.error('You must be logged in to use `arigato run`');
          return resolve(false);
        }

        reject(err);
      });
    });
  }

  spawn (params) {
    return new Promise((resolve, reject) => {
      var segments = params.cmd.split(' ');
      var proc = child_process.spawn(segments[0], segments.slice(1), {
        cwd: process.cwd(),
        env: _.extend(process.env, params.credentials),
        detached: false
      });

      // TODO: Come back and pipe stdin to the child process properly
      proc.stdout.pipe(process.stdout);
      proc.stderr.pipe(process.stderr);

      // TODO: Come back and write proper event management layer for events ->
      // promises (with cleanup)
      function handleSignal(signal) {
        proc.kill(signal);
      }

      var signals = ['SIGHUP', 'SIGINT', 'SIGQUIT', 'SIGTERM'];
      signals.forEach((signal) => {
        process.on(signal, handleSignal.bind(this, signal));
      });

      proc.on('error', (err) => {
        proc.removeAllListeners();
        reject(err);
      });

      proc.on('close', (exitCode) => {
        proc.removeAllListeners();
        resolve(exitCode === null ? exitCode : 1);
      });
    });
  }
}

module.exports = {
  key: 'run',
  synopsis: 'runs a shell command and injects credentials into the environment',
  usage: 'arigato run <shell command here>',
  Command: Run,
  example: `\tlocalhost$ arigato run "bin/api --PORT=8080"`
};
