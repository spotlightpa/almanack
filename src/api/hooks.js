import Vue from "vue";
import VueCompositionAPI from "@vue/composition-api";

Vue.use(VueCompositionAPI);

export { useAuth } from "./auth.js";
export { useAPI } from "./store.js";
