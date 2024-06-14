# SLO Document

This document describes the SLOs for my-app service.

| | |
| --- | --- |
| **Author** | <no value> |


## SLO: foo-availability

| | |
| --- | --- |
| **Compliance Period** | 4w |

### SLI Implementation

| | |
| --- | --- |
| **Source** | prometheus |

```
sum by (job) (rate(http_requests_total{job="foo", code!~"2.."}[4w])) / sum by (job) (rate(http_requests_total{job="foo"}[4w]))
```

### SLO Target

<no value>

### Clarification and Caveats

<no value>


## SLO: bar-availability

| | |
| --- | --- |
| **Compliance Period** | 4w |

### SLI Implementation

| | |
| --- | --- |
| **Source** | prometheus |

```
sum by (job) (rate(http_requests_total{job="bar", code!~"2.."}[4w])) / sum by (job) (rate(http_requests_total{job="bar"}[4w]))
```

### SLO Target

<no value>

### Clarification and Caveats

<no value>


