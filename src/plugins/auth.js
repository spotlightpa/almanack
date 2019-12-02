import Vue from "vue";
import TheAuth from "../components/TheAuth.vue";

let authComponent = new Vue(TheAuth);

export let AuthPlugin = {
  install(Vue) {
    Vue.prototype.$auth = authComponent;
  }
};

export function authGuard(to, from, next) {
  let path = "/login" + window.location.hash;
  authComponent.isSignedIn ? next() : next({ path });
}
