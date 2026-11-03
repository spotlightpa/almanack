import { apdate, aptime } from "journalize";

type DateInput = Date | string | null | undefined;

function toDate(d: DateInput): Date {
  return typeof d === "string" ? new Date(d) : (d as Date);
}

const toWeekday = new Intl.DateTimeFormat("en-US", {
  weekday: "long",
});

export function formatDate(d: DateInput): string {
  if (!d) {
    return "";
  }
  const date = toDate(d);
  return toWeekday.format(date) + ", " + apdate(date);
}

const tzNameLookup = new Intl.DateTimeFormat("en-US", {
  timeZoneName: "short",
});

function getTimeZoneName(d: Date): string {
  const { value = "" } =
    tzNameLookup
      .formatToParts(d)
      .find((part) => part.type === "timeZoneName") ?? {};
  return value;
}

export function formatTime(d: DateInput): string {
  const date = toDate(d);
  let tzname = getTimeZoneName(date);
  if (tzname) {
    tzname = " " + tzname;
  }
  return aptime(date) + tzname;
}

const toShortWeekday = new Intl.DateTimeFormat("en-US", {
  weekday: "short",
});

export function formatDateTime(d: DateInput): string {
  if (!d) {
    return "";
  }
  const date = toDate(d);
  const tz = getTimeZoneName(date);
  const tzSuffix = tz ? " " + tz : "";
  return (
    aptime(date) +
    " " +
    toShortWeekday.format(date) +
    "., " +
    apdate(date) +
    tzSuffix
  );
}

export function today(): Date {
  const d = new Date();
  d.setHours(d.getHours() + 1);
  d.setMinutes(0);
  d.setSeconds(0);
  return d;
}

export function tomorrow(): Date {
  const d = new Date();
  d.setDate(d.getDate() + 1);
  d.setHours(5);
  d.setMinutes(0);
  d.setSeconds(0);
  return d;
}
