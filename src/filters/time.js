import Vue from "vue";

import { apdate, aptime } from "journalize";

let dateOpts = {
  weekday: "short",
};

const dateLocalizer = new Intl.DateTimeFormat("en-US", dateOpts);

export function dateFormatter(d) {
  if (typeof d === "string") {
    d = new Date(d);
  }
  return dateLocalizer.format(d) + "., " + apdate(d) + ", " + aptime(d);
}

Vue.filter("formatDate", dateFormatter);
