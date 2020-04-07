import Vue from "vue";
import router from "./router.js";
import "./filters/index.js";
import "./plugins/font-awesome.js";
import "./plugins/buefy.js";

import VueMeta from "vue-meta";
Vue.use(VueMeta);

import App from "./components/TheApp.vue";

Vue.config.ignoredElements = ["raw-html"];

new Vue({
  router,
  render: (h) => h(App),
}).$mount("#app");
