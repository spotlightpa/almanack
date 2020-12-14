import * as Sentry from "@sentry/browser";

let dsn =
  process.env.NODE_ENV === "production"
    ? "https://cf41d56053f841ae9625673c3ab8d53f@o361657.ingest.sentry.io/3944373"
    : "";

Sentry.init({ dsn });

import { createApp } from "vue";
import router from "./router.js";
import registerFA from "./plugins/font-awesome.js";
import registerBuefy from "./plugins/buefy.js";
// import VueMeta from "vue-meta";

import App from "./components/TheApp.vue";

// Vue.config.ignoredElements = ["raw-html"];

let app = createApp(App);
registerFA(app);
registerBuefy(app);
app.use(router);
app.mount("#app");
