# PiKV

<img width="256px" src="logo.png" alt="logo">

![GitHub release](https://img.shields.io/github/tag/joway/pikv.svg?label=release)
[![Go Report Card](https://goreportcard.com/badge/github.com/joway/pikv)](https://goreportcard.com/report/github.com/joway/pikv)
[![CircleCI](https://circleci.com/gh/joway/pikv.svg?style=shield)](https://circleci.com/gh/joway/imagic)
[![codecov](https://codecov.io/gh/joway/pikv/branch/master/graph/badge.svg)](https://codecov.io/gh/joway/pikv)

A redis protocol compatible key-value store. It's built on top of [Badger](https://github.com/dgraph-io/badger).

## Feature

- [x] Master-Slave Architecture
- [ ] Slave of with key prefix

## Install

```bash
$ go get -u github.com/joway/pikv


# check version
$ pikv -v

# run pikv server
$ pikv -p 6380 --dataDir /data

# connect pikv server
$ redis-cli -p 6380
```

## Benchmark

### pikv with ssd storage
 
```bash
redis-benchmark -p 6380 -q -t SET,GET -P 1024 -r 1000000000 -n 10000000
SET: 89827.09 requests per second
GET: 473619.41 requests per second
```

### pikv with memory storage

```bash
redis-benchmark -p 6380 -q -t SET,GET -P 1024 -r 1000000000 -n 10000000
SET: 943930.56 requests per second
GET: 2215821.00 requests per second
```

### redis

```bash
redis-benchmark -p 6379 -q -t SET,GET -P 1024 -r 1000000000 -n 10000000
SET: 318582.94 requests per second
GET: 604887.50 requests per second
```

### Environment:

- MacBook Pro (13-inch, 2018, Four Thunderbolt 3 Ports)
- 2.3 GHz Quad-Core Intel Core i5
- 16 GB 2133 MHz LPDDR3
