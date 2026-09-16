from redis.cluster import RedisCluster, ClusterNode

startup_nodes = [
    ClusterNode("redis-cluster-0.redis-cluster.svc.cluster.local", 6379),
    ClusterNode("redis-cluster-1.redis-cluster.svc.cluster.local", 6379),
    ClusterNode("redis-cluster-2.redis-cluster.svc.cluster.local", 6379),
]

rc = RedisCluster(startup_nodes=startup_nodes, decode_responses=True)

# set with TTL
rc.set("user:123", json.dumps(user_obj), ex=300)

# get
val = rc.get("user:123")
if val:
    user_obj = json.loads(val)

# remove
rc.delete("user:123")
