import { apdate, aptime } from "journalize";

const toWeekday = new Intl.DateTimeFormat("en-US", {
  weekday: "long",
});

export function formatDate(d) {
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

function getTimeZoneName(d) {
  let { value = "" } = tzNameLookup
    .formatToParts(d)
    .find((part) => part.type === "timeZoneName");
  return value;
}

export function formatTime(d) {
  if (typeof d === "string") {
    d = new Date(d);
  }
  let tzname = getTimeZoneName(d);
  if (tzname) {
    tzname = " " + tzname;
  }
  return aptime(d) + tzname;
}

const toShortWeekday = new Intl.DateTimeFormat("en-US", {
  weekday: "short",
});

export function formatDateTime(d) {
  if (!d) {
    return "";
  }
  if (typeof d === "string") {
    d = new Date(d);
  }
  let tz = getTimeZoneName(d);
  tz = tz && " " + tz;
  return aptime(d) + " " + toShortWeekday.format(d) + "., " + apdate(d) + tz;
}

export function today() {
  let d = new Date();
  d.setHours(d.getHours() + 1);
  d.setMinutes(0);
  d.setSeconds(0);
  return d;
}

export function tomorrow() {
  let d = new Date();
  d.setDate(d.getDate() + 1);
  d.setHours(5);
  d.setMinutes(0);
  d.setSeconds(0);
  return d;
}
