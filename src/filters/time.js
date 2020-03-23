import Vue from "vue";

import { apdate, aptime } from "journalize";

const toWeekday = new Intl.DateTimeFormat("en-US", {
  weekday: "long",
});

export function dateFormatter(d) {
  if (!d) {
    return "";
  }
  if (typeof d === "string") {
    d = new Date(d);
  }
  return toWeekday.format(d) + ", " + apdate(d);
}

const tzNameLookup = new Intl.DateTimeFormat("en-US", {
  timeZoneName: "short",
});

export function timeFormatter(d) {
  if (typeof d === "string") {
    d = new Date(d);
  }
  let { value: tzname } = tzNameLookup
    .formatToParts(d)
    .find((part) => part.type === "timeZoneName") ?? { value: "" };
  if (tzname) {
    tzname = " " + tzname;
  }
  return aptime(d) + tzname;
}

Vue.filter("formatDate", dateFormatter);
Vue.filter("formatTime", timeFormatter);
