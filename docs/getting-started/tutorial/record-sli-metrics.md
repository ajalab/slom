# Record SLI metrics

This document describes how to write an SLO spec to record SLI metrics of the [`example` service](./prerequisite.md).

## Create a spec to record the error rate

The below example is a minimum SLO spec file that records the error rate of the HTTP requests from `http_requests_total` metrics.
This is equivalent to the SLI for an availability SLO.

```yaml title="example.yaml", linenums="1"
--8<-- "examples/tutorial/spec/sli.yaml"
```

The `name` field contains the name of the SLO spec.
An SLO spec file corresponds to a single service or critical user journey, so we set the service name `example` to this field.

The `slos` field contains the list of SLO declarations. Here, we first declare a single SLO `availability` for this service.

The `indicator` contains the configurations for the service level indicator (SLI).
As we are using Prometheus to monitor our service, we write the indicator configurations under `prometheus` field.

The `errorRatio` field in the `prometheus` indicator specifies a [PromQL](https://prometheus.io/docs/prometheus/latest/querying/basics/) query to calculate the ratio of errors to all requests.
The range for the `rate` operator is left unspecified using the `$window` placeholder.
This allows slom to generate rules for deriving error rates over different time windows.
Additionally, this query retains the `job` label using `sum by`.
This is because Prometheus may be monitoring other services as well.

The `level` field in the `prometheus` indicator specifies the [aggregation level](https://prometheus.io/docs/practices/rules/#naming) of the query. We set `[job]` to the field, since we retain `job` label as described above.

The `windows` field specifies a list of time windows for recording the SLI.
For now, specify only a 5-minute rolling window to record the current error rate. Later, we will add longer time windows for recording error budgets and issuing alerts.

## Generate a Prometheus rule file

Run [`slom generate prometheus-rule`](../../references/cli/generate/prometheus_rule.md) command to generate a Prometheus rule file based on the SLO spec.

```sh
slom generate prometheus-rule example.yaml
```

The following output will be displayed.

```yaml
--8<-- "examples/tutorial/out/prometheus-rule-prometheus/sli.yaml"
```

This is a Prometheus rule file that evaluates [recording rules](https://prometheus.io/docs/prometheus/latest/configuration/recording_rules/) or alerting rules.
You can find recording rules that

- record the 5-minute error rate metric as `job:slom_error:ratio_rate5m`
- record the metadata for the SLO as `slom_slo`.

For more details about the generated Prometheus rules, please refer to the [Prometheus](../../references/metrics/prometheus/) reference.
