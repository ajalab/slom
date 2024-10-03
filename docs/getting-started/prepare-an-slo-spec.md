# Prepare an SLO spec

This is a complete example of slom SLO spec, which specifies a 99% availability SLO based on a metric `http_requests_total`.
This spec also defines alerting rules based on the SLO.

- Alerts on burn rate. The below example adopts the [multiwindow, multi-burn-rate alerting](https://sre.google/workbook/alerting-on-slos/#6-multiwindow-multi-burn-rate-alerts) approach for two different time windows and error rate thresholds.
- Alerts on SLO breach.

To learn how to write an SLO spec, please refer to the [Tutorial](../guides/tutorial/).

```yaml title="example.yaml"
--8<-- "examples/getting_started/spec/example.yaml"
```
