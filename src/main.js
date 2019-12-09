import Vue from "vue";

import { library } from "@fortawesome/fontawesome-svg-core";
import {
  faCopy,
  faFileCode,
  faFileWord,
  faNewspaper
} from "@fortawesome/free-regular-svg-icons";
import { faFileDownload, faLink } from "@fortawesome/free-solid-svg-icons";
import { FontAwesomeIcon } from "@fortawesome/vue-fontawesome";

import App from "./components/TheApp.vue";
import router from "./router.js";
import { AuthPlugin } from "./plugins/auth.js";
import { APIPlugin } from "./plugins/api.js";

import "./filters/index.js";

library.add(
  faCopy,
  faFileCode,
  faFileWord,
  faNewspaper,
  faFileDownload,
  faLink
);

Vue.component("font-awesome-icon", FontAwesomeIcon);
Vue.use(AuthPlugin);
Vue.use(APIPlugin);

Vue.config.ignoredElements = ["raw-html"];

new Vue({
  router,
  render: h => h(App)
}).$mount("#app");
