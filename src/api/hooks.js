import Vue from "vue";
import VueCompositionAPI from "@vue/composition-api";

Vue.use(VueCompositionAPI);

import { makeAuth } from "./auth.js";
import { makeClient } from "./service.js";
import { listAvailable, available } from "./arc-services.js";
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

export function useFeed() {
  let feed = listAvailable(useClient());
  feed.initLoad();
  return feed;
}

export function getAvailableArticle() {
  return available(useClient());
}

export function useScheduler(id) {
  let obj = getScheduledArticle({ service: useClient(), id });
  obj.initLoad();
  return obj;
}
