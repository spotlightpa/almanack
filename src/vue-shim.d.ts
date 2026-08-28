declare module "*.vue" {
  import type { DefineComponent } from "vue";
  const component: DefineComponent;
  export default component;
}

declare module "journalize" {
  export { apdate, apdatetab, apmonth, apmonthtab, apnumber, aptime, capfirst, intcomma, intword, ordinal, ordinalsuffix, pluralize, widont, yesno } from "../node_modules/journalize/types/index.d.ts";
}
