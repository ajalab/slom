# Prerequisite

This tutorial assumes a basic understanding of the concept of [service level objectives](https://sre.google/sre-book/service-level-objectives/) (SLOs) and [alerts using SLOs](https://sre.google/workbook/alerting-on-slos/).

## Settings

Throughout this tutorial, we aim to set up SLOs for a service `example`. The `example` service is a simple web service with an HTTP API endpoint. This service is monitored by Prometheus, which collects a metric `http_requests_total` from it.

```
# HELP http_requests_total The total number of HTTP requests.
# TYPE http_requests_total counter
http_requests_total{code="200"} 999
http_requests_total{code="500"} 1
```

Prometheus monitors the service through a job `example`, so the metrics include a label `job="example"`.
