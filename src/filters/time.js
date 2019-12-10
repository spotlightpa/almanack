import Vue from "vue";

import { apdate } from "journalize";

const toWeekday = new Intl.DateTimeFormat("en-US", {
  weekday: "long",
});

export function dateFormatter(d) {
  if (typeof d === "string") {
    d = new Date(d);
  }
  return toWeekday.format(d) + ", " + apdate(d);
}

Vue.filter("formatDate", dateFormatter);
