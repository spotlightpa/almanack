import Vue from "vue";
import Router from "vue-router";
import { watch } from "@vue/composition-api";

import { useAuth } from "@/api/hooks.js";
import { setDimensions, sendGAPageview } from "@/utils/google-analytics.js";

import ViewAdmin from "./components/ViewAdmin.vue";
import ViewArticleItem from "./components/ViewArticleItem.vue";
import ViewArticleList from "./components/ViewArticleList.vue";
import ViewArticleSchedule from "./components/ViewArticleSchedule.vue";
import ViewAuthorizedDomains from "./components/ViewAuthorizedDomains.vue";
import ViewError from "./components/ViewError.vue";
import ViewLogin from "./components/ViewLogin.vue";
import ViewUploader from "./components/ViewUploader.vue";

Vue.use(Router);

let router = new Router({
  mode: "history",
  base: process.env.BASE_URL,
  routes: [
    { path: "/", name: "home", redirect: { name: "articles" } },
    {
      path: "/login",
      name: "login",
      component: ViewLogin,
      meta: {
        title: "Spotlight PA Almanack - Login",
      },
    },
    {
      path: "/articles",
      name: "articles",
      component: ViewArticleList,
      meta: {
        requiresAuth: true,
        title: "Spotlight PA Almanack - List",
      },
    },
    {
      path: "/articles/:id",
      name: "article",
      component: ViewArticleItem,
      props: true,
      meta: { requiresAuth: true },
    },
    {
      path: "/articles/:id/schedule",
      name: "schedule",
      component: ViewArticleSchedule,
      props: true,
      meta: {
        requiresAuth: true,
        title: "Spotlight PA Almanack - Scheduler",
      },
    },
    {
      path: "/admin",
      name: "admin",
      component: ViewAdmin,
      meta: {
        requiresAuth: true,
        title: "Spotlight PA Almanack - Admin",
      },
    },
    {
      path: "/admin/uploader",
      name: "uploader",
      component: ViewUploader,
      meta: {
        requiresAuth: true,
        title: "Spotlight PA Almanack - Image Uploader",
      },
    },
    {
      path: "/admin/domains",
      name: "domains",
      component: ViewAuthorizedDomains,
      meta: {
        requiresAuth: true,
        title: "Spotlight PA Almanack - Preapproved Domains",
      },
    },
    {
      path: "/*",
      name: "error",
      component: ViewError,
    },
  ],
  scrollBehavior(to, from, savedPosition) {
    return savedPosition || { selector: "#top-nav" };
  },
});

let { roles, fullName, email, isSignedIn } = useAuth();

router.beforeEach((to, from, next) => {
  if (to?.meta?.title) {
    document.title = to.meta.title;
  }

  if (to.matched.some((record) => record.meta.requiresAuth)) {
    if (!isSignedIn.value) {
      next({
        name: "login",
        hash: to.hash, // For verifying tokens etc.
      });
      return;
    }
  }
  next();
});

router.afterEach((to) => {
  let domain = /@(.*)$/.exec(email.value)?.[1].toLowerCase() ?? "None";
  let name = fullName.value || "Not Signed In";
  let role =
    roles.value.find((r) => r === "admin") ||
    roles.value.find((r) => r === "Spotlight PA") ||
    roles.value.find((r) => r === "editor") ||
    "None";

  setDimensions({
    domain,
    name,
    role,
  });
  sendGAPageview(to.fullPath);
});

watch(
  () => isSignedIn.value,
  (newStatus, oldStatus) => {
    if (oldStatus === undefined || newStatus === oldStatus) {
      return;
    }
    let name = newStatus ? "articles" : "login";
    router.push({ name });
  }
);

export default router;
