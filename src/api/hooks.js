import Vue from "vue";
import VueCompositionAPI from "@vue/composition-api";

Vue.use(VueCompositionAPI);

import { makeAuth } from "./auth.js";
import { makeService } from "./service.js";
import { listAvailable, available } from "./arc-services.js";
import { getScheduledArticle } from "./scheduler.js";

let $auth;

export function useAuth() {
  if (!$auth) {
    $auth = makeAuth();
  }
  return $auth;
}

let $service;

export function useService() {
  if (!$service) {
    $service = makeService(useAuth());
  }
  return $service;
}

export function useFeed() {
  let feed = listAvailable(useService());
  feed.initLoad();
  return feed;
}

export function getAvailableArticle() {
  return available(useService());
}

export function useScheduler(id) {
  let obj = getScheduledArticle({ service: useService(), id });
  obj.initLoad();
  return obj;
}
