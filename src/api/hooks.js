import Vue from "vue";
import VueCompositionAPI from "@vue/composition-api";

Vue.use(VueCompositionAPI);

import { makeAuth } from "./auth.js";
import { makeClient } from "./service.js";
import { listAvailable, getAvailable, upcoming } from "./arc-services.js";
import { getScheduledArticle } from "./scheduler.js";

let $auth;

export function useAuth() {
  if (!$auth) {
    $auth = makeAuth();
  }
  return $auth;
}

let $client;

export function useClient() {
  if (!$client) {
    $client = makeClient(useAuth());
  }
  return $client;
}

export function useAvailableList() {
  let list = listAvailable(useClient());
  list.load();
  return list;
}

export function getAvailableArticle(id) {
  let loader = getAvailable({ client: useClient(), id });
  loader.load();
  return loader;
}

export function useUpcoming() {
  let client = useClient();
  let loader = upcoming(client);
  loader.loadAndRefresh();
  return loader;
}
export function useScheduler(id) {
  let obj = getScheduledArticle({ client: useClient(), id });
  obj.load();
  return obj;
}
