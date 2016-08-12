'use strict';

var gulp = require('gulp');
var gulpIf = require('gulp-if');
var mocha = require('gulp-mocha');
var eslint = require('gulp-eslint');
var runSequence = require('run-sequence');
var gulpNSP = require('gulp-nsp');

var LINT_FILES = ['./**/*.js', '!node_modules'];

function isFixed(file) {
  return file.eslint && file.eslint.fixed;
}

gulp.task('default', ['lint', 'mocha']);

gulp.task('test', function () {
  runSequence('nsp', 'lint', 'mocha');
});

gulp.task('lint', function () {
  return gulp.src(LINT_FILES)
    .pipe(eslint())
    .pipe(eslint.format())
    .pipe(eslint.failAfterError());
});

gulp.task('fmt', function () {
  return gulp.src(LINT_FILES)
    .pipe(eslint({
      fix: true
    }))
    .pipe(eslint.format())
    .pipe(gulpIf(isFixed, gulp.dest('./')));
});

gulp.task('mocha', function () {
  return gulp.src('./tests/**/*.js', { read: false })
    .pipe(mocha({ reporter: 'spec' }));
});

gulp.task('nsp', function (cb) {
  gulpNSP({
    stopOnError: false, // We'll triage these notifications manually
    package: __dirname + '/package.json'
  }, cb);
});
