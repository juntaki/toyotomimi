# toyotomimi

[![Build Status](https://travis-ci.org/juntaki/toyotomimi.svg?branch=master)](https://travis-ci.org/juntaki/toyotomimi)
[![Coverage Status](https://coveralls.io/repos/github/juntaki/toyotomimi/badge.svg?branch=master)](https://coveralls.io/github/juntaki/toyotomimi?branch=master)
[![GoDoc](https://godoc.org/github.com/juntaki/toyotomimi/lib?status.svg)](https://godoc.org/github.com/juntaki/toyotomimi/lib)

Internet radio recorder written in golang. :radio:

Radiko.jpと、らじる★らじるの番組を全部録音します。

## How to use

Install

~~~
 go get github.com/juntaki/toyotomimi
~~~

Run command.

~~~
$ toyotomimi outputDir
 ...


$ ls outputDir
[2017-0717-1400][ラジオ第1]ニュース.m4a
[2017-0717-1400][ラジオ第2]Japan & World Update.m4a
 ...
~~~

## Requirements

swfextract
