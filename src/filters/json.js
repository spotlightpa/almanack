import Vue from "vue";

function jsonFormatter(o) {
  return JSON.stringify(o, null, 2);
}

Vue.filter("json", jsonFormatter);
