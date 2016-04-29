'use strict';

var gulp = require('gulp');
var jshint = require('gulp-jshint');
var mocha = require('gulp-mocha');

gulp.task('test', ['lint', 'mocha']);
gulp.task('default', ['test']);

gulp.task('mocha', function () {
  return gulp.src('./tests/**/*.js', { read: false })
    .pipe(mocha({ reporter: 'spec' }));
});

gulp.task('lint', function() {
  return gulp.src(['./lib/**/*.js', './tests/**/*.js', 'Gulpfile.js'])
    .pipe(jshint())
    .pipe(jshint.reporter('default'));
});
