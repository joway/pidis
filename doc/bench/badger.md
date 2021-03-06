## Benchmark with Badger

This benchmark test raw set/get.

### pidis with badger storage
 
```bash
redis-benchmark -p 6380 -q -t SET,GET -P 1024 -r 1000000000 -n 10000000
SET: 89827.09 requests per second
GET: 473619.41 requests per second
```

### pidis with memory storage

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
