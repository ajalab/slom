# Alert on error budget consumption

slom supports specifying alerting rules that trigger when a certain portion of the error budget has been consumed within the SLO window.


## Update the spec to alert on error budget consumption

Suppose we want to trigger an alert when 90% of the error budget has been consumed within the current four-week rolling SLO window.
To achieve this, update the [previous SLO spec](./alert-on-error-budget-burn-rate.md) to include a new alert with `errorBudget` field.

```yaml title="example.yaml"
--8<-- "examples/tutorial/spec/alert_error_budget.yaml"
```

1. UPDATED: Added `errorBudget` alert

## Generate a Prometheus rule file

After updating the SLO spec file, run [`slom generate prometheus-rule`](../../references/cli/generate/prometheus_rule.md) command to generate a Prometheus rule file based on the SLO spec.

```shell
slom generate prometheus-rule example.yaml
```

Then, the following output will be displayed.

```yaml
--8<-- "examples/tutorial/out/prometheus-rule-prometheus/alert_error_budget.yaml"
```

You can find that there is a new alerting rule `SLOTooMuchErrorBudgetConsumed`, which is triggered when 90% of the error budget has been consumed in the current SLO window.
