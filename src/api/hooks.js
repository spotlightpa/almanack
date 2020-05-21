import Vue from "vue";
import VueCompositionAPI from "@vue/composition-api";
Vue.use(VueCompositionAPI);

export { useAuth } from "./auth.js";

export { useClient } from "./service.js";
export { makeState } from "./service-util.js";

export {
  useAvailableList,
  useUpcoming,
  getAvailableArticle,
} from "./arc-services.js";

export { useScheduler } from "./scheduler.js";
