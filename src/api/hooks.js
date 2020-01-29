import Vue from "vue";
import VueCompositionAPI from "@vue/composition-api";

Vue.use(VueCompositionAPI);

import { makeAuth } from "./auth.js";
import { makeService } from "./service.js";
import { makeFeed } from "./feed.js";
import { getScheduledArticle } from "./scheduler.js";

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

export function useFeed() {
  let feed = makeFeed(useService());
  feed.initLoad();
  return feed;
}

export function useScheduler(id) {
  let obj = getScheduledArticle({ service: useService(), id });
  obj.initLoad();
  return obj;
}
