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
errorRatio: sum by (job) (rate(http_requests_total{job="foo", code!~"2.."}[$window])) / sum by (job) (rate(http_requests_total{job="foo"}[$window]))
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
errorRatio: sum by (job) (rate(http_requests_total{job="bar", code!~"2.."}[$window])) / sum by (job) (rate(http_requests_total{job="bar"}[$window]))
```

### SLO Target

<no value>

### Clarification and Caveats

<no value>


