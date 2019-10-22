# Pidis

<img width="256px" src="logo.png" alt="logo">

![GitHub release](https://img.shields.io/github/tag/joway/pidis.svg?label=release)
[![Go Report Card](https://goreportcard.com/badge/github.com/joway/pidis)](https://goreportcard.com/report/github.com/joway/pidis)
[![CircleCI](https://circleci.com/gh/joway/pidis.svg?style=shield)](https://circleci.com/gh/joway/pidis)
[![codecov](https://codecov.io/gh/joway/pidis/branch/master/graph/badge.svg)](https://codecov.io/gh/joway/pidis)

A redis protocol compatible key-value store. It's built on top of [Badger](https://github.com/dgraph-io/badger).

## Install

### From docker

```bash
docker run \
  --rm \
  --name=pidis \
  -p 6380:6380 \
  -p 6381:6381 \
  -v /tmp/pidis:/data \
  joway/pidis:latest \
  -p 6380 --rpcPort 6381 -d /data
```

### From go get

```bash
$ go get -u github.com/joway/pidis/...

# check version
$ pidis -v

# run pidis server
$ pidis -p 6380 -d /data

# connect pidis server
$ redis-cli -p 6380
```

## Supported Redis Keys

- KV
  - GET key  
  - SET key value [EX seconds|PX milliseconds] [NX|XX]
  - SETNX key value  
  - DEL key [key ...]
  - EXISTS key [key ...]
  - INCR key
  - TTL key

## Benchmark

### Environment

- CircleCI Docker
- medium

### Pidis

```bash
$ redis-benchmark -p 6380 -d 1000 -q -t SET,GET -P 1024 -r 1000000000 -n 1000000

SET: 13641.26 requests per second
GET: 243546.03 requests per second
```

### LedisDB

```bash
$ redis-benchmark -p 6380 -d 1000 -q -t SET,GET -P 1024 -r 1000000000 -n 1000000

SET: 22494.66 requests per second
GET: 75058.17 requests per second
```

### Redis

```bash
$ redis-benchmark -p 6379 -d 1000 -q -t SET,GET -P 1024 -r 1000000000 -n 1000000

SET: 365497.06 requests per second
GET: 448028.66 requests per second
```
