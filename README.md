# sqlread

[![GoDoc](https://godoc.org/github.com/donatj/sqlread?status.svg)](https://godoc.org/github.com/donatj/sqlread)
[![CI](https://github.com/donatj/sqlread/actions/workflows/ci.yml/badge.svg)](https://github.com/donatj/sqlread/actions/workflows/ci.yml)

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

## User query grammar

Square brackets (`[]`) mark optional clauses. Braces (`{}`) require one choice.
A vertical bar (`|`) separates choices. An ellipsis (`...`) marks a repeated item.

```
SELECT { * | column_name } [, { * | column_name }] ...
    FROM table_name
    [INTO OUTFILE string]
    [LIMIT { row_count | offset, row_count }];

SHOW TABLES;
SHOW COLUMNS FROM table_name;
SHOW CREATE TABLE table_name;
{ EXIT | QUIT };
```

Keywords are case-insensitive. Table names and column names may use backticks.
`string` values may use single or double quotes. `offset` and `row_count` must be
non-negative integers. Expressions, `WHERE`, and `ORDER BY` are not supported.

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

> SELECT building_id, title FROM buildings LIMIT 2;
2,Home Building
190,Test Building (demo)
2017/12/22 12:24:44 restarting lexer

> SELECT `building_id`, `title` FROM `buildings` INTO OUTFILE "dump.csv";
2017/12/22 12:27:43 written to `dump.csv`
2017/12/22 12:27:43 restarting lexer
```
