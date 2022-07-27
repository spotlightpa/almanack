import * as Sentry from "@sentry/browser";

let dsn =
  process.env.NODE_ENV === "production"
    ? "https://cf41d56053f841ae9625673c3ab8d53f@o361657.ingest.sentry.io/3944373"
    : "";

Sentry.init({ dsn });

import Vue from "vue";

import VueMeta from "vue-meta";
Vue.use(VueMeta);

import fontAwesome from "./plugins/font-awesome.js";
Vue.use(fontAwesome);

import project from "./plugins/autoimport.js";
Vue.use(project);

import Router from "vue-router";
Vue.use(Router);

import TheApp from "./components/TheApp.vue";
import router from "./plugins/router.js";

Vue.config.ignoredElements = ["raw-html"];

new Vue({
  router,
  render: (h) => h(TheApp),
}).$mount("#app");
