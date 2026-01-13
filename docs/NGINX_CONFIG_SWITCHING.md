# Switching Between Nginx Configurations

This guide explains how to switch between development and production nginx configurations for testing.

## Quick Switch (Recommended)

### Switch to Production Config
```bash
task nginx-production
```

This will:
1. Copy `nginx.production.conf` to `nginx.conf`
2. Test the configuration
3. Reload nginx (no downtime)

### Switch Back to Development Config
```bash
task nginx-dev
```

This will:
1. Restore `nginx.conf` from git
2. Test the configuration
3. Reload nginx (no downtime)

## Manual Method

### Option 1: Copy the file
```bash
# Switch to production
cp config/nginx/nginx.production.conf config/nginx/nginx.conf
docker-compose exec nginx nginx -t
docker-compose exec nginx nginx -s reload

# Switch back to dev
git checkout config/nginx/nginx.conf
docker-compose exec nginx nginx -t
docker-compose exec nginx nginx -s reload
```

### Option 2: Use docker-compose override
```bash
# Start with production config
docker-compose -f docker-compose.yml -f docker-compose.production.yml up -d nginx

# Start with dev config (default)
docker-compose up -d nginx
```

## Testing Production Config

1. **Switch to production config**:
   ```bash
   task nginx-production
   ```

2. **Verify it's working**:
   ```bash
   curl http://localhost:8080/nginx-health
   ```

3. **Run load test**:
   ```bash
   # Make sure backend servers are running
   ./scripts/start-servers.sh 4
   
   # Run load test against nginx (port 8080)
   API_PORT=8080 task test-load ARGS="-users=1000 -duration=60s -read-only -warmup-writes=1000"
   ```

4. **Compare with direct backend**:
   ```bash
   # Test against single backend (bypass nginx)
   API_PORT=8081 task test-load ARGS="-users=1000 -duration=60s -read-only -warmup-writes=1000"
   ```

## Configuration Differences

### Development (`nginx.conf`)
- Basic configuration
- Lower connection limits
- No rate limiting
- Simpler logging

### Production (`nginx.production.conf`)
- Optimized for high throughput
- Higher connection limits (10k per worker)
- Rate limiting per endpoint
- Connection limiting (DDoS protection)
- Gzip compression
- Enhanced logging with timing metrics
- Keep-alive optimizations

## Troubleshooting

### Config test fails
```bash
# Check nginx error logs
docker-compose logs nginx

# Test config manually
docker-compose exec nginx nginx -t
```

### Changes not applying
```bash
# Restart nginx container
task nginx-restart
# or
docker-compose restart nginx
```

### Want to see both configs side-by-side
```bash
diff config/nginx/nginx.conf config/nginx/nginx.production.conf
```

## Notes

- The production config includes rate limiting (which you already have at app level)
- You can modify `nginx.production.conf` to remove rate limiting if desired
- Both configs use the same upstream backend configuration
- Production config is optimized for 1000+ RPS
