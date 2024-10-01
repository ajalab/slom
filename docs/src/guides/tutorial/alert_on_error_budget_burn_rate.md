# Alert on error budget burn rate

It is generally recommended to trigger an alert when a certain portion of the SLO error budget is consumed, in order to prevent the budget from being fully depleted.
Google's [SRE Workbook](https://sre.google/workbook/table-of-contents/) introduces the concept of [_burn rate_](https://sre.google/workbook/alerting-on-slos/#4-alert-on-burn-rate) as a technique to implement alerting mechanisms like this.

Slogen also supports specifying a Prometheus alerting rule to trigger alerts for SLO error budget consumption.

## Alert on single burn rate

Suppose we want to trigger a page when 1% of the error budget is consumed within an hour.
This corresponds to setting an alert for a burn rate of 2% * 28 days / 1 hour = 13.44.

To generate a Prometheus rule file for such alert, we update the [previous SLO spec](./record_error_budget_metrics.md) like below.

```yaml
# example.yaml
{{#include ../../../../examples/tutorial/spec/alert_single_burn_rate.yaml}}
```

The `alerts` field contains alert specifications.

The `name` field in `alerts` items specifies the name of the alert.

The `burnRate` field in `alerts` items signifies a burn rate-based alert. We set `consumedBudgetRatio` to 0.02, so that an alert is triggered when 2% of the error budget is consumed within the window. Also, we set the window period to 1 hour by specifying `window-1h` in the `singleWindow` field.

The `alerter` field in `alerts` items defines how alerts are triggered. In this case, we use the `prometheus` alerter, which allows you to configure the alert name, labels and annotations as described in the [official guide](https://prometheus.io/docs/prometheus/latest/configuration/alerting_rules/).

After updating the SLO spec file, run [`slogen generate prometheus-rule`](../../references/cli/generate/prometheus_rule.md) command to generate a Prometheus rule file based on the SLO spec.

```shell
slogen generate prometheus-rule example.yaml
```

Then, the following output will be displayed.

```yaml
{{#include ../../../../examples/tutorial/out/prometheus-rule-prometheus/alert_single_burn_rate.yaml}}
```

You can find that there is a new alerting rule `SLOHighErrorRate`, which is triggered when the burn rate reaches 13.44.

## Alert on multiple burn rates

It is a good idea to trigger alerts over different time windows to capture two kinds of issues: _fast burn_ and _slow burn_.

- Fast burn: Alerts are triggered when a significant portion the of error budget is consumed in a short time window (e.g., one hour). This detects urgent issues and typically pages someone for immediate attention.
- Slow burn: Alerts are triggered when a portion of the error budget is consumed over a longer window (e.g., three days). This detects issues that tend to go unnoticed but can gradually exhaust the error budget. A ticket is usually filed for resolution during regular working hours.

You can configure alerts on multiple burn rates by adding more items to the `alerts` field.
The example code below configures alerts for:

- Fast burn: Triggered when 2% of the error budget is consumed within 1 hour (burn rate = 2% * 28 days / 1 hour = 13.44).
- Slow burn: Triggered when 10% of the error budget is consumed within 3 days (burn rate = 10% * 28 days / 3 days = 0.933).

```yaml
# example.yaml
{{#include ../../../../examples/tutorial/spec/alert_multi_burn_rates.yaml}}
```

After running [`slogen generate prometheus-rule`](../../references/cli/generate/prometheus_rule.md) for the updated spec file, you can find that a new alerting rule for slow burn is added.

```yaml
{{#include ../../../../examples/tutorial/out/prometheus-rule-prometheus/alert_multi_burn_rates.yaml}}
```

## Alert with multiple windows

In [6: Multiwindow, Multi-Burn-Rate Alerts](https://sre.google/workbook/alerting-on-slos/#6-multiwindow-multi-burn-rate-alerts) section of Google's [SRE Workbook](https://sre.google/workbook/table-of-contents/), it is recommended to combine the burn rate alerting rule with a shorter window (e.g., 1/12 of the original window). This approach reduces the alert reset time and minimizes the number of false positives.

The code below provides an updated example that configures:

- A 5-minute window for the fast burn alert (1 hour).
- A 6-hour window for the slow burn alert (3 days).

```yaml
# example.yaml
{{#include ../../../../examples/tutorial/spec/alert_multi_windows.yaml}}
```

You will notice that the updated alert specifications utilize `multiWindows` instead of `singleWindow` to configure short time windows.

After running [`slogen generate prometheus-rule`](../../references/cli/generate/prometheus_rule.md) for the updated spec file, you can find that the alerting rules recognize short windows.
