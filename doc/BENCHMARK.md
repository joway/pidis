## Environment

- m4.xlarge (4 cpu, 16 GB)
- 100GB gp2 EBS 

## LedisDB

```bash
$ redis-benchmark -p 6380 -q -t SET,GET -P 1024 -r 1000000000 -n 1000000

SET: 22494.66 requests per second
GET: 75058.17 requests per second
```


