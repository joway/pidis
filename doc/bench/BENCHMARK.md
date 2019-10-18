## Environment

- m4.xlarge (4 cpu, 16 GB)
- 100GB gp2 EBS 

## PiKV

```bash
$ redis-benchmark -p 6380 -q -t SET,GET -P 1024 -r 1000000000 -n 1000000

SET: 13641.26 requests per second
GET: 243546.03 requests per second
```

## LedisDB

```bash
$ redis-benchmark -p 6380 -q -t SET,GET -P 1024 -r 1000000000 -n 1000000

SET: 22494.66 requests per second
GET: 75058.17 requests per second
```

## Redis

```bash
$ redis-benchmark -p 6379 -q -t SET,GET -P 1024 -r 1000000000 -n 1000000

SET: 365497.06 requests per second
GET: 448028.66 requests per second
```
