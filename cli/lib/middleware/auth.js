'use strict';

// stub middleware for passthrough generation.
function auth() {
  return true;
}

module.exports = function () {
  return auth;
};
