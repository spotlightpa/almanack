import { createApp } from "vue";
import TheApp from "./components/TheApp.vue";
const app = createApp(TheApp);

import * as Sentry from "@sentry/vue";

if (import.meta.env.MODE === "production") {
  let dsn =
    "https://cf41d56053f841ae9625673c3ab8d53f@o361657.ingest.sentry.io/3944373";
  Sentry.init({ app, dsn });
}

import { Head, createHead } from "@vueuse/head";
const head = createHead();
app.use(head);
app.component("MetaHead", Head);

import fontAwesome from "./plugins/font-awesome.js";
app.use(fontAwesome);

import router from "./plugins/router.js";
app.use(router);
router.app = app;

app.mount("#app");
