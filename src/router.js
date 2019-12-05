import Vue from "vue";
import Router from "vue-router";
import ViewArticleItem from "./components/ViewArticleItem.vue";
import ViewArticleList from "./components/ViewArticleList.vue";
import ViewError from "./components/ViewError.vue";
import ViewLogin from "./components/ViewLogin.vue";

import { authGuard } from "./plugins/auth.js";

Vue.use(Router);

export default new Router({
  mode: "history",
  base: process.env.BASE_URL,
  routes: [
    { path: "/", name: "home", redirect: { name: "articles" } },
    {
      path: "/login",
      name: "login",
      component: ViewLogin
    },
    {
      path: "/articles",
      name: "articles",
      component: ViewArticleList,
      beforeEnter: authGuard
    },
    {
      path: "/articles/:id",
      name: "article",
      component: ViewArticleItem,
      props: true,
      beforeEnter: authGuard
    },
    {
      path: "/*",
      name: "error",
      component: ViewError
    }
  ],
  scrollBehavior(to, from, savedPosition) {
    return savedPosition || { selector: "#top-nav" };
  }
});
