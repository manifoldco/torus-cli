'use strict';

let pad = exports;

/**
 * Pads a string on the right side by adding the char until the string
 * is at its max length.
 *
 * If the str exceeds the max len then its sliced.
 *
 * A space is used by default
 */
pad.right = function (str, maxLen, char) {
  char = char || ' ';
  str = (str.length > maxLen) ? str.slice(0, maxLen) : str;

  var distance = maxLen - str.length;
  for (var i = distance; i > 0; i--) {
    str += char;
  }

  return str;
};
