'use strict';

var gulp = require('gulp');
var mocha = require('gulp-mocha');
var eslint = require('gulp-eslint');
var runSequence = require('run-sequence');

gulp.task('default', ['lint', 'mocha']);

gulp.task('test', function () {
  runSequence('lint', 'mocha');
});

gulp.task('lint', function () {
  return gulp.src(['./lib/**/*.js', './tests/**/*.js', 'Gulpfile.js'])
    .pipe(eslint({ extends: 'airbnb-base/legacy' }))
    .pipe(eslint.format())
    .pipe(eslint.failAfterError());
});

gulp.task('mocha', function () {
  return gulp.src('./tests/**/*.js', { read: false })
    .pipe(mocha({ reporter: 'spec' }));
});
