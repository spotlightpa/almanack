import * as Sentry from "@sentry/browser";

let dsn =
  import.meta.env.NODE_ENV === "production"
    ? "https://cf41d56053f841ae9625673c3ab8d53f@o361657.ingest.sentry.io/3944373"
    : "";

Sentry.init({ dsn });

import { createApp } from "vue";
import TheApp from "./components/TheApp.vue";

const app = createApp(TheApp);

import { createMetaManager } from "vue-meta";

app.use(createMetaManager());
// app.use(metaPlugin);

import fontAwesome from "./plugins/font-awesome.js";
app.use(fontAwesome);

import project from "./plugins/autoimport.js";
app.use(project);

import router from "./plugins/router.js";

app.use(router);

app.mount("#app");
