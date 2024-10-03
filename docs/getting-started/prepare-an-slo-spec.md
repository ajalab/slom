# Prepare an SLO spec

The following is a complete example of a slom SLO specification, which defines a 99% availability SLO using the `http_requests_total` metric. This specification also includes alerting rules based on the SLO:

- It triggers alerts when the error budget burn rate surpasses defined thresholds. The example uses the [multiwindow, multi-burn-rate alerting](https://sre.google/workbook/alerting-on-slos/#6-multiwindow-multi-burn-rate-alerts) strategy, applying different thresholds for two time windows.
- It triggers alerts when the error budget consumption exceeds a specified threshold within the SLO window.

```yaml title="example.yaml"
--8<-- "examples/getting_started/spec/example.yaml"
```

For detailed guidance on writing an SLO specification, please refer to the [Tutorial](../guides/tutorial/).
