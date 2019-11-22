import Vue from "vue";
import Router from "vue-router";
import ViewLogin from "./components/ViewLogin.vue";
import ViewHome from "./components/ViewHome.vue";

import { authGuard } from "./plugins/auth.js";

Vue.use(Router);

export default new Router({
  mode: "history",
  base: process.env.BASE_URL,
  routes: [
    { path: "/", redirect: "/home" },
    {
      path: "/login",
      name: "login",
      component: ViewLogin
    },
    {
      path: "/home",
      name: "home",
      component: ViewHome,
      beforeEnter: authGuard
    }
  ]
});
