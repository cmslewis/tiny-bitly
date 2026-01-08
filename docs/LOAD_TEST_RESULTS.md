
# Load Test Results

## V1

- Rate limiting disabled
- 1 req/s
- No composite index
- No connection pooling

### Concurrent writes

| Metric | 100 users | 1000 | 10k | 25k | 50k | 100k |
| --- | --- | --- | --- | --- | --- | --- |
| P95 | 60ms | 150ms | 4.72s | 29.7s | 25.4s | 42.6s |
| Requests Failed | 0% | 0% | 0% | 65.91% | 99.9% | 100% |

<details>
<summary>100 users</summary>
<pre>
Duration:           30s
Total Requests:     3000
Successful:         3000 (100.00%)
Failed:             0 (0.00%)
Throughput:         99.80 req/s

Error Breakdown:
  Rate Limited (429): 0
  Timeouts:           0
  Server Errors (5xx): 0
  Client Errors (4xx): 0

Request Type Breakdown:
  Writes: 3000 total, 3000 successful (100.00%), 0 rate limited, 0 client errors

Latency Statistics (successful requests only):
  Min:    3.265ms
  P50:    44.636ms
  P75:    51.582ms
  P90:    56.515ms
  P95:    59.949ms
  P99:    66.327ms
  P99.9:  71.089ms
  Max:    72.702ms
  Avg:    39.841ms
</pre>
</details>

<details>
<summary>1000 users</summary>
<pre>
Duration:           30s
Total Requests:     30000
Successful:         30000 (100.00%)
Failed:             0 (0.00%)
Throughput:         994.14 req/s

Error Breakdown:
  Rate Limited (429): 0
  Timeouts:           0
  Server Errors (5xx): 0
  Client Errors (4xx): 0

Request Type Breakdown:
  Writes: 30000 total, 30000 successful (100.00%), 0 rate limited, 0 client errors

Latency Statistics (successful requests only):
  Min:    8.778ms
  P50:    102.195ms
  P75:    124.502ms
  P90:    140.764ms
  P95:    149.807ms
  P99:    184.906ms
  P99.9:  211.135ms
  Max:    220.278ms
  Avg:    101.945ms
</pre>
</details>

<details>
<summary>10k users</summary>
<pre>
Duration:           31s
Total Requests:     190919
Successful:         190919 (100.00%)
Failed:             0 (0.00%)
Throughput:         6160.12 req/s

Error Breakdown:
  Rate Limited (429): 0
  Timeouts:           0
  Server Errors (5xx): 0
  Client Errors (4xx): 0

Request Type Breakdown:
  Writes: 190919 total, 190919 successful (100.00%), 0 rate limited, 0 client errors

Latency Statistics (successful requests only):
  Min:    4.248ms
  P50:    818.663ms
  P75:    1.655765s
  P90:    2.977264s
  P95:    4.71832s
  P99:    8.261908s
  P99.9:  10.178806s
  Max:    16.225098s
  Avg:    1.346782s
</pre>
</details>

<details>
<summary>25k users</summary>
<pre>
Duration:           54s
Total Requests:     72266
Successful:         24634 (34.09%)
Failed:             47632 (65.91%)
Throughput:         1338.14 req/s

Error Breakdown:
  Rate Limited (429): 0
  Timeouts:           0
  Server Errors (5xx): 588
  Client Errors (4xx): 0

Request Type Breakdown:
  Writes: 72266 total, 24634 successful (34.09%), 0 rate limited, 0 client errors

Latency Statistics (successful requests only):
  Min:    9.598ms
  P50:    22.116377s
  P75:    24.855161s
  P90:    29.258059s
  P95:    29.69443s
  P99:    29.878376s
  P99.9:  29.967747s
  Max:    29.992456s
  Avg:    20.022327s

Performance Indicators:
  ⚠️  Success rate is below 95% - system may be overloaded
</pre>
</details>

<details>
<summary>50k users</summary>
<pre>
Duration:           1m1s
Total Requests:     71765
Successful:         41 (0.06%)
Failed:             71724 (99.94%)
Throughput:         1184.82 req/s

Error Breakdown:
  Rate Limited (429): 0
  Timeouts:           0
  Server Errors (5xx): 0
  Client Errors (4xx): 0

Request Type Breakdown:
  Writes: 71765 total, 41 successful (0.06%), 0 rate limited, 0 client errors

Latency Statistics (successful requests only):
  Min:    22.512738s
  P50:    24.972069s
  P75:    25.251197s
  P90:    25.350797s
  P95:    25.417571s
  P99:    25.43149s
  P99.9:  25.43149s
  Max:    25.43149s
  Avg:    24.474433s

Performance Indicators:
  ⚠️  Success rate is below 95% - system may be overloaded
</pre>
</details>

<details>
<summary>100k users</summary>
<pre>
Duration:           1m14s
Total Requests:     226903
Successful:         2 (0.00%)
Failed:             226901 (100.00%)
Throughput:         3080.43 req/s

Error Breakdown:
  Rate Limited (429): 0
  Timeouts:           0
  Server Errors (5xx): 0
  Client Errors (4xx): 0

Request Type Breakdown:
  Writes: 226903 total, 2 successful (0.00%), 0 rate limited, 0 client errors

Latency Statistics (successful requests only):
  Min:    23.457199s
  P50:    42.641681s
  P75:    42.641681s
  P90:    42.641681s
  P95:    42.641681s
  P99:    42.641681s
  P99.9:  42.641681s
  Max:    42.641681s
  Avg:    33.04944s

Performance Indicators:
  ⚠️  Success rate is below 95% - system may be overloaded
</pre>
</details>


### Concurrent reads

| Metric | 100 users | 1000 | 10k | 25k | 50k | 100k |
| --- | --- | --- | --- | --- | --- | --- |
| P95 | 60ms | 3.48s | 29.6s |  |  |  |
| Requests Failed | 0% | 61.17% | 98.0% | | |

<details>
<summary>100 users</summary>
<pre>
Duration:           30s
Total Requests:     3000
Successful:         3000 (100.00%)
Failed:             0 (0.00%)
Throughput:         99.75 req/s

Error Breakdown:
  Rate Limited (429): 0
  Timeouts:           0
  Server Errors (5xx): 0
  Client Errors (4xx): 0

Request Type Breakdown:
  Reads:  3000 total, 3000 successful (100.00%), 0 rate limited, 0 client errors

Latency Statistics (successful requests only):
  Min:    4.083ms
  P50:    61.164ms
  P75:    70.892ms
  P90:    80.016ms
  P95:    83.576ms
  P99:    92.195ms
  P99.9:  98.148ms
  Max:    99.612ms
  Avg:    57.439ms
</pre>
</details>

<details>
<summary>1000 users</summary>
<pre>
Duration:           31s
Total Requests:     24019
Successful:         9326 (38.83%)
Failed:             14693 (61.17%)
Throughput:         771.40 req/s

Error Breakdown:
  Rate Limited (429): 0
  Timeouts:           0
  Server Errors (5xx): 7233
  Client Errors (4xx): 0

Request Type Breakdown:
  Reads:  24019 total, 9326 successful (38.83%), 0 rate limited, 0 client errors

Latency Statistics (successful requests only):
  Min:    697µs
  P50:    478.494ms
  P75:    860.991ms
  P90:    1.607137s
  P95:    3.476756s
  P99:    11.263432s
  P99.9:  17.888468s
  Max:    29.116497s
  Avg:    880.981ms

Performance Indicators:
  ⚠️  Success rate is below 95% - system may be overloaded
  ⚠️  High server error rate (30.11%) - check application logs
</pre>
</details>

<details>
<summary>10k users</summary>
<pre>
Duration:           58s
Total Requests:     15374
Successful:         312 (2.03%)
Failed:             15062 (97.97%)
Throughput:         263.47 req/s

Error Breakdown:
  Rate Limited (429): 0
  Timeouts:           0
  Server Errors (5xx): 7218
  Client Errors (4xx): 0

Request Type Breakdown:
  Reads:  15374 total, 312 successful (2.03%), 0 rate limited, 0 client errors

Latency Statistics (successful requests only):
  Min:    49.358ms
  P50:    24.355744s
  P75:    27.719269s
  P90:    29.36702s
  P95:    29.579266s
  P99:    29.865529s
  P99.9:  29.888026s
  Max:    29.888026s
  Avg:    20.761957s

Performance Indicators:
  ⚠️  Success rate is below 95% - system may be overloaded
  ⚠️  High server error rate (46.95%) - check application logs
</pre>
</details>

<details>
<summary>25k users</summary>
<pre>

</pre>
</details>

<details>
<summary>50k users</summary>
<pre>

</pre>
</details>

<details>
<summary>100k users</summary>
<pre>

</pre>
</details>

## V2

- Added composite index to speed up queries that filter on short code AND expires_at

### Concurrent reads

| Metric | 100 users | 1000 | 10k | 25k | 50k | 100k |
| --- | --- | --- | --- | --- | --- | --- |
| (Before) P95 | 60ms | 3.48s | 29.6s |  |  |  |
| (After) P95 | 81.4ms | 2.50s | 29.5s |  |  |  |
| (Before) Requests Failed | 0% | 61.17% | 98.0% | | |
| (After) Requests Failed | 0% | 62.2% | 97.6% |  |  |  |

<details>
<summary>100 users</summary>
<pre>
Duration:           30s
Total Requests:     3000
Successful:         3000 (100.00%)
Failed:             0 (0.00%)
Throughput:         99.73 req/s

Error Breakdown:
  Rate Limited (429): 0
  Timeouts:           0
  Server Errors (5xx): 0
  Client Errors (4xx): 0

Request Type Breakdown:
  Reads:  3000 total, 3000 successful (100.00%), 0 rate limited, 0 client errors

Latency Statistics (successful requests only):
  Min:    2.392ms
  P50:    62.3ms
  P75:    70.62ms
  P90:    77.72ms
  P95:    81.44ms
  P99:    92.711ms
  P99.9:  96.439ms
  Max:    96.671ms
  Avg:    59.29ms
</pre>
</details>

<details>
<summary>1000 users</summary>
<pre>
Duration:           31s
Total Requests:     24504
Successful:         9256 (37.77%)
Failed:             15248 (62.23%)
Throughput:         785.98 req/s

Error Breakdown:
  Rate Limited (429): 0
  Timeouts:           0
  Server Errors (5xx): 7265
  Client Errors (4xx): 0

Request Type Breakdown:
  Reads:  24504 total, 9256 successful (37.77%), 0 rate limited, 0 client errors

Latency Statistics (successful requests only):
  Min:    659µs
  P50:    464.673ms
  P75:    737.247ms
  P90:    1.633989s
  P95:    2.50481s
  P99:    10.571144s
  P99.9:  13.529554s
  Max:    13.534188s
  Avg:    800.93ms

Performance Indicators:
  ⚠️  Success rate is below 95% - system may be overloaded
  ⚠️  High server error rate (29.65%) - check application logs
</pre>
</details>

<details>
<summary>10000 users</summary>
<pre>
Duration:           1m0s
Total Requests:     16265
Successful:         386 (2.37%)
Failed:             15879 (97.63%)
Throughput:         271.98 req/s

Error Breakdown:
  Rate Limited (429): 0
  Timeouts:           0
  Server Errors (5xx): 8859
  Client Errors (4xx): 0

Request Type Breakdown:
  Reads:  16265 total, 386 successful (2.37%), 0 rate limited, 0 client errors

Latency Statistics (successful requests only):
  Min:    88.4ms
  P50:    26.433034s
  P75:    27.86001s
  P90:    28.765409s
  P95:    29.516769s
  P99:    29.953952s
  P99.9:  29.989798s
  Max:    29.989798s
  Avg:    24.259443s

Performance Indicators:
  ⚠️  Success rate is below 95% - system may be overloaded
  ⚠️  High server error rate (54.47%) - check application logs
</pre>
</details>


## V3

- Added PostgreSQL connection pooling (PgBouncer)

### Concurrent reads

| Metric | 100 users | 1000 | 10k | 25k | 50k | 100k |
| --- | --- | --- | --- | --- | --- | --- |
| (Before) P95 | 81.4ms | 2.50s | 29.5s | - |  |  |
| (Before) Requests Failed | 0% | 62.2% | 97.6% | - |  |  |
| (After) P95 | 68.8ms | 681ms | 24.6s | 24.2s |  |  |
| (After) Requests Failed | 0% | 0% | 27.3% | 73.5% |  |  |

<details>
<summary>100 users</summary>
<pre>
Duration:           30s
Total Requests:     3000
Successful:         3000 (100.00%)
Failed:             0 (0.00%)
Throughput:         99.77 req/s

Error Breakdown:
  Rate Limited (429): 0
  Timeouts:           0
  Server Errors (5xx): 0
  Client Errors (4xx): 0

Request Type Breakdown:
  Reads:  3000 total, 3000 successful (100.00%), 0 rate limited, 0 client errors

Latency Statistics (successful requests only):
  Min:    8.166ms
  P50:    31.767ms
  P75:    52.531ms
  P90:    63.517ms
  P95:    68.832ms
  P99:    80.348ms
  P99.9:  90.145ms
  Max:    90.331ms
  Avg:    37.875ms
</pre>
</details>

<details>
<summary>1000 users</summary>
<pre>
Duration:           31s
Total Requests:     30000
Successful:         30000 (100.00%)
Failed:             0 (0.00%)
Throughput:         977.42 req/s

Error Breakdown:
  Rate Limited (429): 0
  Timeouts:           0
  Server Errors (5xx): 0
  Client Errors (4xx): 0

Request Type Breakdown:
  Reads:  30000 total, 30000 successful (100.00%), 0 rate limited, 0 client errors

Latency Statistics (successful requests only):
  Min:    5.591ms
  P50:    221.285ms
  P75:    615.178ms
  P90:    661.538ms
  P95:    681.122ms
  P99:    703.547ms
  P99.9:  734.724ms
  Max:    739.871ms
  Avg:    365.485ms
</pre>
</details>

<details>
<summary>10000 users</summary>
<pre>
Duration:           37s
Total Requests:     43094
Successful:         31324 (72.69%)
Failed:             11770 (27.31%)
Throughput:         1170.72 req/s

Error Breakdown:
  Rate Limited (429): 0
  Timeouts:           0
  Server Errors (5xx): 58
  Client Errors (4xx): 0

Request Type Breakdown:
  Reads:  43094 total, 31324 successful (72.69%), 0 rate limited, 0 client errors

Latency Statistics (successful requests only):
  Min:    7.057ms
  P50:    5.08517s
  P75:    15.941161s
  P90:    21.569563s
  P95:    24.622446s
  P99:    28.482948s
  P99.9:  29.795125s
  Max:    29.979467s
  Avg:    8.391991s

Performance Indicators:
  ⚠️  Success rate is below 95% - system may be overloaded
</pre>
</details>

<details>
<summary>25000 users</summary>
<pre>
Duration:           40s
Total Requests:     64307
Successful:         17030 (26.48%)
Failed:             47277 (73.52%)
Throughput:         1609.29 req/s

Error Breakdown:
  Rate Limited (429): 0
  Timeouts:           0
  Server Errors (5xx): 58
  Client Errors (4xx): 0

Request Type Breakdown:
  Reads:  64307 total, 17030 successful (26.48%), 0 rate limited, 0 client errors

Latency Statistics (successful requests only):
  Min:    13.568ms
  P50:    12.510234s
  P75:    15.220656s
  P90:    18.022777s
  P95:    24.168018s
  P99:    25.132422s
  P99.9:  25.410112s
  Max:    25.512219s
  Avg:    12.71835s

Performance Indicators:
  ⚠️  Success rate is below 95% - system may be overloaded
</pre>
</details>


## V4

- Added Redis cache in front of database

### Concurrent reads

| Metric | 100 users | 1000 | 10k | 25k | 50k | 100k |
| --- | --- | --- | --- | --- | --- | --- |
| (Before Redis) P95 | 68.8ms | 681ms | 24.6s | 24.2s |  |  |
| (Before Redis) Requests Failed | 0% | 0% | 27.3% | 73.5% |  |  |
| (After Redis) P95 | 27.2ms | 519.5ms | 20.89s | 25.9s |  |  |
| (After Redis) Requests Failed | 0% | 0% | 16.2% | 74.1% |  |

**Analysis:**
- **100-1000 users**: Redis provides significant improvement (60% faster at 100 users, 24% faster at 1000 users)
- **10k users**: Redis provides modest improvement (15% faster, 11% fewer failures)
- **25k users**: Redis provides minimal/no improvement (actually 7% slower P95, similar failure rate)

**Conclusion:** Redis helps significantly at moderate concurrency (100-1000 users) but becomes a bottleneck at extreme concurrency (25k+ users). At 25k concurrent users, Redis's single-threaded nature and network overhead outweigh the benefits. The system is fundamentally overloaded at this scale.  |

<details>
<summary>100 users</summary>
<pre>
Duration:           30s
Total Requests:     3000
Successful:         3000 (100.00%)
Failed:             0 (0.00%)
Throughput:         99.91 req/s

Error Breakdown:
  Rate Limited (429): 0
  Timeouts:           0
  Server Errors (5xx): 0
  Client Errors (4xx): 0

Request Type Breakdown:
  Reads:  3000 total, 3000 successful (100.00%), 0 rate limited, 0 client errors

Latency Statistics (successful requests only):
  Min:    5.001ms
  P50:    22.001ms
  P75:    24.504ms
  P90:    26.529ms
  P95:    27.201ms
  P99:    28.804ms
  P99.9:  30.009ms
  Max:    31.838ms
  Avg:    21.625ms
</pre>
</details>

<details>
<summary>1000 users</summary>
<pre>
Duration:           31s
Total Requests:     30000
Successful:         30000 (100.00%)
Failed:             0 (0.00%)
Throughput:         982.50 req/s

Error Breakdown:
  Rate Limited (429): 0
  Timeouts:           0
  Server Errors (5xx): 0
  Client Errors (4xx): 0

Request Type Breakdown:
  Reads:  30000 total, 30000 successful (100.00%), 0 rate limited, 0 client errors

Latency Statistics (successful requests only):
  Min:    6.706ms
  P50:    138.844ms
  P75:    450.615ms
  P90:    505.77ms
  P95:    519.472ms
  P99:    530.007ms
  P99.9:  537.175ms
  Max:    539.654ms
  Avg:    276.467ms
</pre>
</details>

<details>
<summary>10000 users</summary>
<pre>
Duration:           48s
Total Requests:     34583
Successful:         28980 (83.80%)
Failed:             5603 (16.20%)
Throughput:         724.67 req/s

Error Breakdown:
  Rate Limited (429): 0
  Timeouts:           0
  Server Errors (5xx): 237
  Client Errors (4xx): 0

Request Type Breakdown:
  Reads:  34583 total, 28980 successful (83.80%), 0 rate limited, 0 client errors

Latency Statistics (successful requests only):
  Min:    82.803ms
  P50:    6.575743s
  P75:    14.995622s
  P90:    18.981054s
  P95:    20.885221s
  P99:    25.194321s
  P99.9:  28.782217s
  Max:    29.720271s
  Avg:    9.520643s

Performance Indicators:
  ⚠️  Success rate is below 95% - system may be overloaded
</pre>
</details>

<details>
<summary>25000 users</summary>
<pre>
Duration:           54s
Total Requests:     65351
Successful:         16950 (25.94%)
Failed:             48401 (74.06%)
Throughput:         1207.18 req/s

Error Breakdown:
  Rate Limited (429): 0
  Timeouts:           0
  Server Errors (5xx): 35
  Client Errors (4xx): 0

Request Type Breakdown:
  Reads:  65351 total, 16950 successful (25.94%), 0 rate limited, 0 client errors

Latency Statistics (successful requests only):
  Min:    113.819ms
  P50:    16.355997s
  P75:    21.391239s
  P90:    22.693976s
  P95:    25.906221s
  P99:    26.084736s
  P99.9:  26.373341s
  Max:    26.423236s
  Avg:    17.521486s

Performance Indicators:
  ⚠️  Success rate is below 95% - system may be overloaded
</pre>
</details>