import Vue from "vue";
import Router from "vue-router";
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
      path: "/*",
      name: "error",
      component: ViewError
    }
  ]
});
