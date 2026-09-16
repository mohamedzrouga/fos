import httpx

CACHE_URL = "http://localhost:8090"

def cache_get(key: str):
    r = httpx.get(f"{CACHE_URL}/get", params={"key": key})
    if r.status_code == 404:
        return None
    return r.json()["value"]

def cache_set(key: str, value, ttl_seconds: int = 300):
    httpx.post(f"{CACHE_URL}/set", json={"key": key, "value": value, "ttl_seconds": ttl_seconds})
