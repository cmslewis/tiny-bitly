# Web-Scale Architecture Guide (100M DAU = 1000 RPS)

This guide outlines the architecture and optimizations needed to handle 100M Daily Active Users (DAU) with an average of 1000 requests per second.

## Traffic Profile

- **Average RPS**: 1000 requests/second
- **Peak RPS**: 10,000-50,000 requests/second (10-50x spikes)
- **Read/Write Ratio**: 80% reads, 20% writes
- **Read Requests**: ~800 RPS average, ~8,000-40,000 RPS peak
- **Write Requests**: ~200 RPS average, ~2,000-10,000 RPS peak

## Architecture Components

### 1. CDN (Content Delivery Network) - **Critical for Scale**

**Purpose**: Cache redirects at the edge, reduce origin load by 90%+

**Implementation**:
- Use Cloudflare, AWS CloudFront, or Google Cloud CDN
- Cache GET requests (redirects) for 24 hours
- Cache key: `{shortCode}`
- Invalidate on write via CDN API

**Benefits**:
- 90%+ of reads served from edge (sub-10ms latency)
- Origin servers handle only cache misses and writes
- Geographic distribution reduces latency globally

**Cost**: ~$5k-20k/month depending on traffic

### 2. NGINX Load Balancer - **Optimized Configuration**

**Capacity**: Single nginx instance can handle 10k+ RPS easily

**Configuration**: See `config/nginx/nginx.production.conf`

**Key Settings**:
- `worker_connections 10000` - 40k concurrent connections (4 cores)
- `keepalive 500` - Reuse backend connections
- Rate limiting per endpoint type
- Connection limiting to prevent exhaustion

**Scaling**:
- Single instance: Up to 10k RPS
- Multiple instances: Use DNS round-robin or L4 load balancer
- Geographic distribution: Deploy in multiple regions

### 3. Application Servers (Go)

**Scaling Strategy**:
- **Horizontal**: Add more instances behind nginx
- **Per Instance**: 100-200 RPS comfortably
- **Total Needed**: 5-10 instances for 1000 RPS average, 50-100 for peak

**Connection Pools**:
- Database: 100 connections per instance (already configured)
- Redis: 200 connections per instance (already configured)

**Resource Requirements**:
- CPU: 2-4 cores per instance
- Memory: 2-4 GB per instance
- Total: 10-40 instances × 2-4 cores = 20-160 cores

### 4. Redis Cache - **Critical for Performance**

**Purpose**: Cache reads to reduce database load by 95%+

**Configuration**:
- Connection pool: 200 per app instance
- Memory: 20-100 GB (for 1B URLs)
- Clustering: Redis Cluster for high availability

**Cache Strategy**:
- Cache all reads (short code → original URL)
- TTL based on record expiration
- Cache-aside pattern (already implemented)

**Scaling**:
- Single instance: Up to 100k ops/sec
- Cluster: Multiple shards for higher throughput
- Memory: 20-100 GB depending on data size

**Cost**: ~$2k/month (Google Memorystore)

### 5. PostgreSQL Database

**Scaling Strategy**:
- **Read Replicas**: 2-5 read replicas for read scaling
- **Connection Pooling**: PgBouncer (50-100 connections per app instance)
- **Write Instance**: Single primary (can handle 1k+ writes/sec)

**Configuration**:
- `max_connections`: 1000-5000 (with connection pooling)
- Connection pool: 50-100 per app instance via PgBouncer
- Indexes: Already optimized (composite index on short_code + expires_at)

**Scaling**:
- Vertical: Larger instances (more CPU/RAM)
- Horizontal: Read replicas for reads, single primary for writes
- Partitioning: By date or hash for very large datasets

**Cost**: ~$500-1500/month (Cloud SQL)

### 6. Connection Pooling (PgBouncer)

**Purpose**: Reduce database connection overhead

**Configuration**:
- Pool mode: Transaction pooling
- Pool size: 50-100 connections per app instance
- Total connections: 500-1000 (across all instances)

**Benefits**:
- Reduces connection churn
- Allows more app instances with fewer DB connections
- Better resource utilization

## Capacity Planning

### Average Load (1000 RPS)

| Component | Instances | Capacity per Instance | Total Capacity |
|-----------|-----------|----------------------|----------------|
| CDN | N/A | Edge caching | 100k+ RPS |
| NGINX | 1-2 | 10k RPS | 10k-20k RPS |
| App Servers | 5-10 | 100-200 RPS | 500-2000 RPS |
| Redis | 1-3 | 100k ops/sec | 100k-300k ops/sec |
| PostgreSQL | 1 primary + 2 replicas | 1k writes/sec, 5k reads/sec | 1k writes, 15k reads |

### Peak Load (50k RPS)

| Component | Instances | Capacity per Instance | Total Capacity |
|-----------|-----------|----------------------|----------------|
| CDN | N/A | Edge caching | 500k+ RPS |
| NGINX | 5-10 | 10k RPS | 50k-100k RPS |
| App Servers | 50-100 | 100-200 RPS | 5k-20k RPS |
| Redis | 3-5 cluster | 100k ops/sec | 300k-500k ops/sec |
| PostgreSQL | 1 primary + 5 replicas | 1k writes/sec, 5k reads/sec | 1k writes, 25k reads |

## Monitoring & Observability

### Key Metrics

1. **NGINX**:
   - Requests per second
   - Response times (P50, P95, P99)
   - Error rates (4xx, 5xx)
   - Active connections
   - Upstream response times

2. **Application**:
   - Request rate per instance
   - Response times
   - Error rates
   - Database connection pool usage
   - Redis connection pool usage

3. **Database**:
   - Query rate
   - Query latency
   - Connection count
   - Replication lag

4. **Redis**:
   - Operations per second
   - Cache hit rate
   - Memory usage
   - Connection count

### Tools

- **Prometheus**: Metrics collection (already configured)
- **Grafana**: Visualization (already configured)
- **Alerting**: Set up alerts for:
  - Error rate > 1%
  - P95 latency > 500ms
  - Cache hit rate < 90%
  - Database connection pool exhaustion

## Cost Estimate

| Component | Monthly Cost |
|-----------|--------------|
| CDN (Cloudflare/AWS) | $5k-20k |
| Redis (Memorystore) | $2k |
| PostgreSQL (Cloud SQL) | $500-1500 |
| App Servers (Compute) | $1k-5k |
| NGINX Load Balancers | $200-500 |
| DNS | $200-500 |
| **Total** | **~$9k-30k/month** |

## Deployment Strategy

### Development/Staging
- Single instance of each component
- Use `nginx.conf` (current config)

### Production
- Use `nginx.production.conf` (optimized config)
- Multiple app instances (5-10 for average, 50-100 for peak)
- Redis cluster (3-5 nodes)
- PostgreSQL with read replicas (1 primary + 2-5 replicas)
- CDN in front of everything

### Geographic Distribution
- Deploy in multiple regions (US, EU, Asia)
- Use CDN for global distribution
- Database replication across regions (with eventual consistency)

## Performance Targets

- **P95 Latency**: < 100ms (with CDN), < 500ms (without CDN)
- **P99 Latency**: < 200ms (with CDN), < 1s (without CDN)
- **Availability**: 99.99% (52 minutes downtime/year)
- **Cache Hit Rate**: > 90% for reads
- **Error Rate**: < 0.1%

## Next Steps

1. **Deploy CDN**: Cloudflare or AWS CloudFront
2. **Optimize NGINX**: Use `nginx.production.conf`
3. **Add PgBouncer**: For connection pooling
4. **Scale Redis**: Cluster for high availability
5. **Add Read Replicas**: For database read scaling
6. **Implement Monitoring**: Prometheus + Grafana alerts
7. **Load Testing**: Validate at 10k-50k RPS
8. **Geographic Distribution**: Deploy in multiple regions
