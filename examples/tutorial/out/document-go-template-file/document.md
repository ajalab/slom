# SLO Document

This document describes the SLOs for example service.

| | |
| --- | --- |
| **Author** | john.doe |

## SLO: availability

| | |
| --- | --- |
| **Compliance Period** | 4w |

### SLI Implementation

| | |
| --- | --- |
| **Source** | prometheus |

```
errorRatio: sum by (job) (rate(http_requests_total{job="example", code!~"2.."}[$window])) / sum by (job) (rate(http_requests_total{job="example"}[$window]))
```

### SLO Target

99% of requests were served successfully.

### Clarification and Caveats

- Request metrics are measured at the load balancer.
- We only count HTTP 5XX status messages as error codes; everything else is counted as success.
