import Vue from "vue";

let dateOpts = {
  weekday: "short",
  year: "numeric",
  month: "short",
  day: "numeric",
  hour: "numeric",
  minute: "numeric"
};

const dateLocalizer = new Intl.DateTimeFormat("en-US", dateOpts);

function dateFormatter(d) {
  if (typeof d === "string") {
    d = new Date(d);
  }
  return dateLocalizer.format(d);
}

Vue.filter("formatDate", dateFormatter);
