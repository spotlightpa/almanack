import Vue from "vue";

export function commaAndJoiner(a) {
  if (!a || !a.length) {
    return "";
  }
  let ss = a.map((item) => item.toString());
  if (ss.length < 3) {
    return ss.join(" and ");
  }
  let commas = a.slice(0, -1).join(", ");
  return `${commas} and ${ss[ss.length - 1]}`;
}

Vue.filter("commaand", commaAndJoiner);
