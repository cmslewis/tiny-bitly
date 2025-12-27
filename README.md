# tiny-bitly

This is a tiny version of bit.ly built for the purposes of improving my system-design chops.

## Functional requirements

### In scope

1. Users should be able to submit a long URL and receive a shortened one.
    1. Optionally, users should be able to specify a custom alias for their shortened URL.
    1. Optionally, users should be able to specify an expiration date for their shortened URL.
1. Users should be able to access the original URL by using the shortened URL.

### Out of scope

- User authentication and account management
- Analytics on link clicks (e.g., click counts, geographic data)

## Non-functional requirements

Refer to _how_ the system operates, rather than what tasks it performs.

### In scope

- ✅ If the submitted long URL is not a valid URL, a 4xx error is returned.
- ✅ A short URL must correspond to exactly one long URL.
- The redirection should occur with minimal delay (<100ms).
- Must be reliable and available 99.99% of the time (availability > consistency).
- Must support 1B shortened URLs and 100M DAU.

### Out of scope

- Data consistency in real-time analytics
- Advanced security features like spam detection and malicious URL filtering

## Core Entities

- Original URL
- Short URL
- User

## API

- ✅ Shorten a URL:
    ```
    POST /urls
    {
        url: "https://www.example.com/some/very/long/url",
        alias: "optional_alias", // Supports only [A-Za-z0-9_-] to avoid URL-protected characters
        expiresAt: "optional_timestamp"
    }
    ->
    {
        "shortUrl": "https://localhost:3000/abc123"
    }
    ```

- ✅ Access a long URL via a short URL:
    ```
    GET /{short_code}
    -> HTTP 302 Redirect to the original long URL
    ```
    (Both 301 Moved Permanently and 302 Temporary Redirect will redirect a request, but browsers may temporarily cache 301 responses, so 302 is more flexible).

## High-Level Design

### 1. Users should be able to submit a long URL and receive a shortened version

Responsibilities:

1. **Client:** Users interact with the system via a web or mobile application.
2. **Server:** Receives requests from the client and handles all business logic like the short-URL creation and validation.
3. **Database:** Stores the mapping of short codes to long URLs, and user-generated aliases and expiration dates.

Inputs:
```
url: string;
alias?: string;
expiresAt?: string;
```

Outputs:
```
shortUrl: string;
```

Logic:

1. Validate that `url` is a valid URL (using standard Golang package `net/url`).
1. If `alias` provided:
    1. If `alias` already used for another URL (as `short_code`):
        1. If that row's `expires_at` is in the past, delete the row from the DB.
        1. Else, return `400 Bad Request`.
    1. Create row with `short_code=[alias]` (setting `original_url` and `expires_at` if needed).
    1. Return `[SERVICE_URL]/[short_code]`.
1. Until success, up to 10 tries:
    1. Create a random `short_code`.
    1. Create row with this `original_url`, `short_code`, `expires_at`, etc.
    1. If the short code already exists, fail and try again.
1. Return `[SERVICE_URL]/[short_code]`.

SHORT_CODE_LENGTH = 6 (b/c this is greater than 1 billion)
Need DB index for `original_url`
Need DB index for `short_code`

### 2. Users should be able to access the original URL by using the shortened URL

Inputs:
```
GET /{short_code}
```

Outputs:
```
302 Redirect to original_url, or
404 Not Found if not found
```

1. Parse `short_code` out of the URL from the GET request.
1. If `short_code` does not exists in DB, return `404 Not Found` response.
1. If `short_code` has expired, return `404 Not Found` response.
1. Else, `302 Temporary Redirect` to its `original_url`.

Options for redirect status codes:
- `301 Moved Permanently`: Browsers may temporarily cache this redirect mapping, meaning subsequent requests for the same URL might go directly to the long URL, bypassing our server and ignoring `expired_at` or any other modifications we might make.
- `302 Temporary Redirect`: Suggests that the resource has been temporarily relocated to a different URL. Browsers do not cache this response, ensuring future requests to this short URL always go through our server.

### Deep Dives

#### 1. Ensuring URLs are unique

To generate a short URL for a given long URL, generate a 6-character short URL using base62 (A-Z, a-z, 0-9). This allows for >1B unique codes (>56B actually). We can increase the length of short codes to 7 if we want to scale the system.

For each creation request, we will generate a short code and then check if it exists in the DB on insert using PostgreSQL's `INSERT ... ON CONFLICT DO NOTHING`. In this case, we will generate a new short code and then try again up to 10 times, then we'll return a 500 error if we still fail to find a unique code.

#### 2. Ensuring redirects are fast

- If we use a relational DB like Postgres without an index on `short_code`: requires sequential scan through O(100GB) of data assuming O(100 bytes) per row, which would take seconds to minutes depending on disk I/O SSD (on something like Cloud SQL).
    - 100M DAU, assuming uniformly distribution, means ~1000 requests per second if we assume each user accesses 1 URL once. But could spike to many times that in real life. 1 access per day is probably a safe assumption.
    - Postgres `max_connections` (on something like Cloud SQL) is 25 to 262,143, though defaults are 25 to 1000 depending on instance size.
    - With 1000 requests per second and `max_connections=1000`, the DB will suffer from connection-churn overhead (process spawning, authentication, session setup, TCP/IP handshake, memory pressure of 50MB per process).
    - Could use **connection pooling** via PgBouncer or Pgpool-II between server and DB (could maintain pool of 50-100 connections).
    - Could introduce **indices** to avoid sequential scan and reduce query duration to <50ms.
    - With 100 `max_connections` (a standard default for postgres out of the box). Micro instances will use 25-50 in Cloud SQL. Postgres performance degrades significantly beyond 300-400 connections. Each connection could handle 1000ms / 50ms = 20 queries per second; with 100 max connections, we can in theory easily handle the 1000 queries per second.
    - Could use hash index, but this is not recommended since you have to REINDEX to shrink it. Easier to use normal B-Tree index, the default, which provides O(log N) lookup.
    - But if we assume 5 redirects per day per user, then we need to support ~5000 queries per second, which requires more connections. Thus, we need to do better.
    - And if we spike to 100x during a moment, that's ~500k queries in that second.
- Introducing **Redis**, an in-memory key-value store, to reduce query time to 0.001 ms(1e3 times faster!). Requires cache invalidation when the TTL expires (TTL=expires_at). Should use LRU if there is no `expires_at`. Feasible to fit 1B shortened URLs at 20 chars * 1e9 = 20e9 bytes = 20 GB in RAM, plus the original_url key = 100s of GB of RAM
    - Can use **Google Memorystore** for managed Redis, offers 1.4 GB to 58 GB per node, so 2-3 instances should be sufficient.
    - To avoid cache stampedes, use mutexes to ensure only one process regenerates the data.
- We can also leverage **CDN** and **Edge Computing**. Could cache the redirect response on an edge node, e.g. with Cloudflare workers, cached for e.g. 24 hours. Can invalidate CDN cache via an API POST request to CDN API
- Would cost O($10k) per month with Redis and CDN costs (Memorystore is $2k, CDN is $5k to $20k per month with APAC traffic more expensive, Origin Servers are $500 to $1500 per month), DNS is $200 - $500 per month)

#### 3. How can we scale to support 1B shortened URLs and 100M DAU?

1B rows * 500 bytes = 500 GB of data 

What if the DB goes down?
- **Postgres replication**
- **Database backup**

To scale reads more than writes, can introduce microservice architecture to split Read Service from Write Service. Then we can horizontally scale the Read Service. We won't need this if we have a good cache hit rate though, since we can fit on one node.

Can't use the UUID approach here. UUID uses a-f and 0-9 and writes may need to happen 10e6 times per day (10x less than reads) = 100 times per seconds, which should be doable on one machine. Instead of this, could also just generate the hexadecimal URL for the primary key autoincremented ID of the DB row.

Typical Postgres supports 500 - 2000 commits per second. We can **batch inserts** if needed to insert 100-500 rows at a time.

Could use a scheduled cleanup job to delete expired URLs periodically (e.g., daily).

#### 4. How do we ensure 99.99% uptime?

Things to consider:

- Multi-region deployment: Active-active or active-passive
- Database replication: Read replicas + failover (mentioned but not detailed)
- Redis replication: Sentinel or cluster mode
- Health checks: For all components
- Circuit breakers: Fallback to DB if Redis is down
- Monitoring: Metrics, alerts, dashboards
- Incident response: Runbooks, on-call rotation
- Graceful degradation: Serve from DB if cache fails
- Load balancing: Multiple instances behind a load balancer
- Automated failover: For DB and cache


#### 5. Other considerations

- Idempotency: what if the same request is sent twice?
- Race conditions: multiple requests for the same alias
- URL validation: only mentions `net/url`, should validate scheme, length, etc.
- Security: Rate limiting, input sanitization, malicious URL detection
- Data model: No schema definition
- API versioning: not mentioned