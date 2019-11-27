import Vue from "vue";

import { library } from "@fortawesome/fontawesome-svg-core";
import { faNewspaper } from "@fortawesome/free-regular-svg-icons";
import { faExternalLinkAlt } from "@fortawesome/free-solid-svg-icons";
import { FontAwesomeIcon } from "@fortawesome/vue-fontawesome";

import App from "./components/TheApp.vue";
import router from "./router.js";
import { AuthPlugin } from "./plugins/auth.js";
import { APIPlugin } from "./plugins/api.js";

import "./filters/intcomma.js";
import "./filters/json.js";
import "./filters/time.js";

library.add(faNewspaper, faExternalLinkAlt);

Vue.component("font-awesome-icon", FontAwesomeIcon);
Vue.use(AuthPlugin);
Vue.use(APIPlugin);

new Vue({
  router,
  render: h => h(App)
}).$mount("#app");
