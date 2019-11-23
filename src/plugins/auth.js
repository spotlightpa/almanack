import Vue from "vue";
import StoreAuth from "../components/StoreAuth.vue";

let authComponent = new Vue(StoreAuth);

export let AuthPlugin = {
  install(Vue) {
    Vue.prototype.$auth = authComponent;
  }
};

export function authGuard(to, from, next) {
  let path = "/login" + window.location.hash;
  authComponent.isSignedIn ? next() : next({ path });
}
