<!-- Copyright NU Cybernetics. p(DOOM) — research prototype. -->
# Disagreement display

Disagreement is never hidden inside one mean. The forecasts page shows, per compatibility group: every member with its wording, population and date; the distribution of member values; the preferred aggregate and every alternative method; and the disagreement label (IQR on the log-odds scale).

Groups compared side by side, with an explicit warning that sample sizes and quality differ:

- general AI researchers (ESPAI surveys, thousands of respondents, low response rates);
- superforecasters (XPT, dozens, strong short-horizon track record);
- domain experts (XPT, dozens);
- public forecasters (Metaculus, hundreds to thousands, self-selected);
- individual experts (Ord, Carlsmith), shown as single points.

Model-family differences are shown between the external aggregate and the research-mode model, with optimistic, central and pessimistic cases from the sensitivity runs. Reasons for disagreement are listed in `research/forecasts/forecast-inventory.md`: question wording, horizon, conditioning, population selection, framing, and beliefs about alignment difficulty and governance.
