import Vue from "vue";
import VueCompositionAPI from "@vue/composition-api";

Vue.use(VueCompositionAPI);

import { makeAuth } from "./auth.js";
import { makeService } from "./service.js";
import { makeAPI } from "./store.js";

let $auth;

export function useAuth() {
  if (!$auth) {
    $auth = makeAuth();
  }
  return $auth;
}

let $service;

function useService() {
  if (!$service) {
    $service = makeService(useAuth());
  }
  return $service;
}

export function useAPI() {
  return makeAPI(useService());
}
