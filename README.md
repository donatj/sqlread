# sqlread

[![Go Report Card](https://goreportcard.com/badge/github.com/donatj/sqlread)](https://goreportcard.com/report/github.com/donatj/sqlread)
[![GoDoc](https://godoc.org/github.com/donatj/sqlread?status.svg)](https://godoc.org/github.com/donatj/sqlread)
[![Build Status](https://travis-ci.org/donatj/sqlread.svg?branch=master)](https://travis-ci.org/donatj/sqlread)
[![Build Status](https://cloud.drone.io/api/badges/donatj/sqlread/status.svg)](https://cloud.drone.io/donatj/sqlread)

sqlread is a mysql dump parser that allows you to query MySQL dumps directly without loading them into MySQL, and export specific tables from said dumps.

Currently, very picky and only likes `mysqldump` generated output dumps. Mileage may vary on dumps created with other tools such as Navicat. Compatibility is a work in progress.

## Todo:

While the code is quite usable and useful, I still have a ways to go.

- [ ] General Expression Support
  - [ ] Basic WHERE support
  - [ ] Basic ORDER BY support
- [ ] Optimization of heavily `\`'d string parsing

## Installation

### From Source

```bash
go install github.com/donatj/sqlread/cmd/sqlread@latest
```

### Precompiled Binaries

Available on the [Releases](https://github.com/donatj/sqlread/releases) page currently for Linux and macOS.

## Example usage

```
$ sqlread buildings.sql
2017/12/22 12:23:52 starting initial pass
2017/12/22 12:23:52 loaded from cache
2017/12/22 12:23:52 finished initial pass
> SHOW TABLES;
buildings
2017/12/22 12:24:01 restarting lexer

> SHOW COLUMNS FROM buildings;
building_id,int
account_id,int
title,varchar
descr,varchar
city,varchar
state_id,smallint
zip_code,varchar
deleted,tinyint
2017/12/22 12:24:08 restarting lexer

> SELECT building_id, title FROM buildings;
2,Home Building
190,Test Building (demo)
192,Donat Building
194,Other Building
201,Sample Building (demo)
205,Johnson Building
2017/12/22 12:24:44 restarting lexer

> SELECT `building_id`, `title` FROM `buildings` INTO OUTFILE "dump.csv";
2017/12/22 12:27:43 written to `dump.csv`
2017/12/22 12:27:43 restarting lexer
```
