# PiKV

<img width="256px" src="logo.png" alt="logo">

![GitHub release](https://img.shields.io/github/tag/joway/pikv.svg?label=release)
[![Go Report Card](https://goreportcard.com/badge/github.com/joway/pikv)](https://goreportcard.com/report/github.com/joway/pikv)
[![CircleCI](https://circleci.com/gh/joway/pikv.svg?style=shield)](https://circleci.com/gh/joway/imagic)
[![codecov](https://codecov.io/gh/joway/pikv/branch/master/graph/badge.svg)](https://codecov.io/gh/joway/pikv)

A redis protocol compatible key-value store. It's built on top of [Badger](https://github.com/dgraph-io/badger).

## TODO

- [x] Master-Slave Architecture
- [ ] Config with toml file
- [ ] Benchmark between redis,ledisdb,pikv
- [ ] Slave of with key prefix
- [ ] ~100% test coverage

## Get Start

### From docker

```bash
docker run \
  --name=pikv
  -p 6380:6380 \
  -p 6381:6381 \
  -v /tmp/pikv:/data \
  joway/pikv:latest \
  -p 6380 --rpcPort 6381 -d /data
```

### From go get

```bash
$ go get -u github.com/joway/pikv/...

# check version
$ pikv -v

# run pikv server
$ pikv -p 6380 -d /data

# connect pikv server
$ redis-cli -p 6380
```

## Supported Redis Keys

- KV
  - GET key  
  - SET key val  
  - DEL key [key ...]

## Benchmark

### Environment

- m4.xlarge (4 cpu, 16 GB)
- 100GB gp2 EBS 

### PiKV

```bash
$ redis-benchmark -p 6380 -q -t SET,GET -P 1024 -r 1000000000 -n 1000000

SET: 13641.26 requests per second
GET: 243546.03 requests per second
```

### LedisDB

```bash
$ redis-benchmark -p 6380 -q -t SET,GET -P 1024 -r 1000000000 -n 1000000

SET: 22494.66 requests per second
GET: 75058.17 requests per second
```

### Redis

```bash
$ redis-benchmark -p 6379 -q -t SET,GET -P 1024 -r 1000000000 -n 1000000

SET: 365497.06 requests per second
GET: 448028.66 requests per second
```
