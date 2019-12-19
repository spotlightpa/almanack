import Vue from "vue";
import Router from "vue-router";
import ViewArticleItem from "./components/ViewArticleItem.vue";
import ViewArticleList from "./components/ViewArticleList.vue";
import ViewError from "./components/ViewError.vue";
import ViewLogin from "./components/ViewLogin.vue";

import { useAuth } from "@/api/hooks.js";
import { watch } from "@vue/composition-api";

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
    },
    {
      path: "/articles",
      name: "articles",
      component: ViewArticleList,
      meta: { requiresAuth: true },
    },
    {
      path: "/articles/:id",
      name: "article",
      component: ViewArticleItem,
      props: true,
      meta: { requiresAuth: true },
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

let { isSignedIn } = useAuth();

router.beforeEach((to, from, next) => {
  if (to.matched.some(record => record.meta.requiresAuth)) {
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
