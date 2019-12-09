import TheAuth from "../components/TheAuth.vue";

export let authComponent;

export let AuthPlugin = {
  install(Vue) {
    let AuthComp = Vue.extend(TheAuth);
    authComponent = new AuthComp();
    Vue.prototype.$auth = authComponent;
  },
};
