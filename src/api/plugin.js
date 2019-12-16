import { $auth } from "./auth.js";
import { createAPIService } from "./service.js";
import { useAPI } from "./store.js";

export default {
  install(Vue) {
    let service = createAPIService();
    Vue.prototype.$auth = $auth;
    Vue.prototype.$api = useAPI(service);
  },
};
