# Record error budget metrics

The _error budget_ for an SLO defines the acceptable amount of unreliability that can occur within a given period without violating the SLO.

This document describes how to write an SLO spec to record remaining error budgets.

## Update the spec to record remaining error budget ratio

We assume that the `example` service has a rolling four-week availability SLO with 99% compliance target.

To record the remaining eror budget, we update the [previous SLO spec](./record_sli_metrics.md) to define `objective` field and a rolling four-week window.

```yaml
# example.yaml
{{#include ../../../../examples/tutorial/spec/error_budget.yaml}}
```

Note that the new `objective` field in the `availability` SLO specifies a `ratio` field with a target compliance ratio 99%, and a `windowRef` field that refers to a rolling four-week window defined in the `windows` field.

## Generate a Prometheus rule file

After updating the SLO spec file, run [`slom generate prometheus-rule`](../../references/cli/generate/prometheus_rule.md) command to generate a Prometheus rule file based on the SLO spec.

```shell
slom generate prometheus-rule example.yaml
```

Then, the following output will be displayed.

```yaml
{{#include ../../../../examples/tutorial/out/prometheus-rule-prometheus/error_budget.yaml}}
```

You can find a new recording rule `job:slom_error_budget:ratio_rate4w` that records the ratio of the remaining error budget to the initial budget (99 - SLO)%.
The value of this metric provides the following insights:

- A value of `1` indicates that no failures have impacted reliability within the time window.
- A positive value means the service is compliant with the SLO at that point in time.
- A negative value signals that the SLO has been breached at that point in time.

For more details about the generated Prometheus rules, please refer to the [Prometheus](../../references/metrics/prometheus/) reference.
