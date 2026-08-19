# Temporal specification

Core temporal primitives distinguish civil/calendar values from real instants.

## Civil date

`temporal.civil_date` represents a Gregorian calendar date without timezone or time-of-day semantics.

Canonical text form is `YYYY-MM-DD`. Years are in the inclusive range `0001..9999`. Invalid month/day combinations are rejected, including non-leap-year February 29.

A civil date must never be converted to an instant without an explicit consumer-owned timezone/context decision.

## Year-month

`temporal.year_month` represents a Gregorian calendar month without timezone semantics.

Canonical text form is `YYYY-MM`, with years in `0001..9999` and months in `01..12`.

## Errors

Temporal validation uses stable semantic codes such as `empty`, `invalid_format`, `invalid_year`, `invalid_month`, `invalid_day`, `zero_value`, and `out_of_range`.
