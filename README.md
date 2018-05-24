# ecocache
Ecocache is a groupcache/redis/memcached like key/value cache store, intended to be used in [specific cases](https://www.cnblogs.com/ecofast/p/9084160.html).</br>
It's inspired by google's [groupcache](https://github.com/golang/groupcache) project and actually, some of ecocache's core code were "copyed" from groupcache such as [consistenthash](https://github.com/golang/groupcache/tree/master/consistenthash) and [lru](https://github.com/golang/groupcache/blob/master/lru/lru.go).</br>

# Architecture
![image](https://github.com/ecofast/ecocache/blob/master/ecocache.png)</br></br>

# Benchmark
![image](https://github.com/ecofast/ecocache/blob/master/cacheserver_win7.jpg)</br>
![image](https://github.com/ecofast/ecocache/blob/master/cacheserver_centos7.png)</br></br>

**Note that this is still early alpha code and needs further development / bug-fix in many place.**
