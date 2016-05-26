'use strict';

module.exports = function () {
  return function (ctx) {
    var name = ctx.program.name;
    var slug = ctx.slug;
    if (!ctx.session) {
      console.log('You must be logged-in to execute \'' +
                  name + ' ' + slug + '\'');
      console.log();
      console.log(
        'Login using \'' + name + ' login\' or create an account with ' +
        '\'' + name + ' signup\'');
      return false;
    }

    return true;
  };
};
