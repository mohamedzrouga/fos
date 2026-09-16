import asyncio
from collections import defaultdict
from typing import Callable, Awaitable, TypeVar, Hashable

T = TypeVar("T")

class SingleFlight:
    def __init__(self):
        self._inflight: dict[Hashable, asyncio.Future] = {}
        self._lock = asyncio.Lock()

    async def do(self, key: Hashable, fn: Callable[[], Awaitable[T]]) -> T:
        async with self._lock:
            fut = self._inflight.get(key)
            if fut is None:
                fut = asyncio.ensure_future(fn())
                self._inflight[key] = fut
            else:
                # someone else already started it; just await
                return await fut

        try:
            return await fut
        finally:
            async with self._lock:
                # only clear if still the same future (avoid races)
                if self._inflight.get(key) is fut:
                    del self._inflight[key]

sf = SingleFlight()

async def fetch_user(user_id: str):
    return await sf.do(f"user:{user_id}", lambda: expensive_fetch(user_id))
