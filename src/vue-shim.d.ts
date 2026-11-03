/// <reference types="vite/client" />

// journalize ships types but has a broken exports field in package.json;
// this shim points TS directly at them.
declare module "journalize" {
  export {
    apdate,
    apdatetab,
    apmonth,
    apmonthtab,
    apnumber,
    aptime,
    capfirst,
    intcomma,
    intword,
    ordinal,
    ordinalsuffix,
    pluralize,
    widont,
    yesno,
  } from "../node_modules/journalize/types/index.d.ts";
}
