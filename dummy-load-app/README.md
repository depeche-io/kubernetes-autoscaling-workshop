Dummy App
===

Generates some CPU, memory and time consumption to mimic some real-life apps for load-testing.

```
# Each request aims for ~100ms total, ~50% of one core busy, allocates ~64MB per request (±20% jitter).
go run main.go -threads=2 -cpu=50 -mem=64 -time=100 -jitter=0.2
```

