import Vue from "vue";
import Router from "vue-router";
import ViewLogin from "./components/ViewLogin.vue";
import ViewArticleList from "./components/ViewArticleList.vue";

import { authGuard } from "./plugins/auth.js";

Vue.use(Router);

export default new Router({
  mode: "history",
  base: process.env.BASE_URL,
  routes: [
    { path: "/", redirect: { name: "home" } },
    {
      path: "/login",
      name: "login",
      component: ViewLogin
    },
    {
      path: "/articles",
      name: "home",
      component: ViewArticleList,
      beforeEnter: authGuard
    }
  ]
});
