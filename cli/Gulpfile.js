'use strict';

var gulp = require('gulp');
var jshint = require('gulp-jshint');
var mocha = require('gulp-mocha');

gulp.task('default', ['lint','mocha']);

gulp.task('mocha', function () {
  return gulp.src('./tests/**/*.js', { read: false })
    .pipe(mocha({ reporter: 'spec' }));
});

gulp.task('lint', function() {
  return gulp.src(['./lib/**/*.js', './tests/**/*.js', 'Gulpfile.js'])
    .pipe(jshint())
    .pipe(jshint.reporter('default'));
});
