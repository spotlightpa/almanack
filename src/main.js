import * as Sentry from "@sentry/browser";

let dsn =
  process.env.NODE_ENV === "production"
    ? "https://cf41d56053f841ae9625673c3ab8d53f@o361657.ingest.sentry.io/3944373"
    : "";

Sentry.init({ dsn });

import Vue from "vue";
import router from "./router.js";
import "./plugins/font-awesome.js";

import VueMeta from "vue-meta";
Vue.use(VueMeta);

import "./plugins/autoimport.js";

import App from "./components/TheApp.vue";

Vue.config.ignoredElements = ["raw-html"];

new Vue({
  router,
  render: (h) => h(App),
}).$mount("#app");
