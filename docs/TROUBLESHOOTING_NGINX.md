# Troubleshooting Nginx Load Testing Issues

## Common Issues and Solutions

### High Failure Rates (4xx/5xx Errors)

#### 1. Check Backend Servers Are Running
```bash
# Verify backend servers are accessible
curl http://localhost:8081/health
curl http://localhost:8082/health
curl http://localhost:8083/health
curl http://localhost:8084/health

# Check if servers are running
./scripts/start-servers.sh 4
```

#### 2. Check Nginx Can Reach Backends
```bash
# Check nginx logs for connection errors
docker-compose logs nginx | grep -i error

# Test from inside nginx container
docker-compose exec nginx wget -O- http://host.docker.internal:8081/health
```

#### 3. Connection Limiting Too Restrictive
If you see many 4xx errors, connection limits might be too low:

**Symptoms:**
- High 4xx error rate (especially 503 Service Unavailable)
- Errors increase with more concurrent users
- "Rate Limited (429): 0" (not rate limiting)

**Solution:**
Increase connection limits in `nginx.production.conf`:
```nginx
limit_conn conn_limit 1000;  # Increase from 100
limit_conn conn_limit 500;   # In location blocks, increase from 50
```

#### 4. Rate Limiting Too Aggressive
If you see 429 errors:

**Symptoms:**
- "Rate Limited (429): X" in load test results
- Errors correlate with request rate

**Solution:**
Increase rate limits or burst sizes:
```nginx
limit_req zone=read_limit burst=1000 nodelay;  # Increase burst
```

### High Latency

#### 1. Backend Overload
**Symptoms:**
- P95 latency > 1s
- High 5xx error rate
- Backend CPU/memory high

**Solution:**
- Add more backend instances
- Check database connection pool
- Check Redis connection pool

#### 2. Nginx Connection Pool Exhaustion
**Symptoms:**
- High upstream connect times in logs
- Timeouts

**Solution:**
Increase keepalive connections:
```nginx
keepalive 1000;  # Increase from 500
```

### Debugging Steps

#### 1. Check Nginx Status
```bash
# View nginx error logs
docker-compose logs -f nginx

# Check nginx config is valid
docker-compose exec nginx nginx -t

# View access logs with timing
docker-compose exec nginx tail -f /var/log/nginx/access.log
```

#### 2. Test Individual Components
```bash
# Test nginx directly
curl http://localhost:8080/nginx-health

# Test backend through nginx
curl http://localhost:8080/health

# Test backend directly (bypass nginx)
curl http://localhost:8081/health
```

#### 3. Monitor Resource Usage
```bash
# Check nginx container resources
docker stats tiny-bitly-nginx-1

# Check backend processes
ps aux | grep "go run.*server"
```

#### 4. Compare With/Without Nginx
```bash
# Test with nginx (port 8080)
API_PORT=8080 task test-load ARGS="-users=1000 -duration=30s"

# Test without nginx (direct backend, port 8081)
API_PORT=8081 task test-load ARGS="-users=1000 -duration=30s"
```

## Quick Fixes for Load Testing

If you're just trying to test performance (not security), temporarily disable limits:

```nginx
# Comment out rate limiting
# limit_req zone=read_limit burst=1000 nodelay;

# Comment out connection limiting
# limit_conn conn_limit 500;
```

Or use the development config which has no limits:
```bash
task nginx-dev
```

## Expected Behavior

### With Production Config (Rate/Connection Limiting)
- Some 4xx errors under extreme load (expected)
- Lower throughput but better protection
- Good for production workloads

### With Development Config (No Limits)
- Higher throughput
- No protection against abuse
- Good for load testing

## Next Steps

1. **Verify backends are running**: `./scripts/start-servers.sh 4`
2. **Check nginx logs**: `docker-compose logs nginx`
3. **Test connectivity**: `curl http://localhost:8080/health`
4. **Try development config**: `task nginx-dev` (no limits)
5. **Compare results**: Test with both configs
