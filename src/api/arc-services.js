import { toRefs, computed } from "@vue/composition-api";

import ArcArticle from "./arc-article.js";
import { makeState } from "./service-util.js";
import { useClient } from "./service.js";

function listAvailable(client) {
  let { apiState, exec } = makeState();

  return {
    ...toRefs(apiState),

    articles: computed(() =>
      apiState.isLoading || apiState.error || !apiState.rawData
        ? []
        : ArcArticle.from(apiState.rawData)
    ),
    load() {
      return exec(() => client.listAvailable());
    },
  };
}

export function useAvailableList() {
  let client = useClient();
  let list = listAvailable(client);
  list.load();
  return list;
}

function getAvailable({ client, id }) {
  let { apiState, exec } = makeState();

  return {
    ...toRefs(apiState),

    article: computed(() =>
      apiState.isLoading || apiState.error || !apiState.rawData
        ? null
        : new ArcArticle(apiState.rawData)
    ),

    canLoad: client.hasAuthAvailable(),
    load() {
      return exec(() => client.getAvailable(id));
    },
  };
}

export function getAvailableArticle(id) {
  let loader = getAvailable({ client: useClient(), id });
  loader.load();
  return loader;
}

function upcoming(client) {
  let { apiState, exec } = makeState();

  return {
    ...toRefs(apiState),

    articles: computed(() =>
      apiState.isLoading || apiState.error || !apiState.rawData
        ? []
        : ArcArticle.from(apiState.rawData)
    ),

    canLoad: client.hasAuthUpcoming(),
    load() {
      return exec(() => client.upcoming());
    },
    loadAndRefresh() {
      return exec(() => client.listRefreshArc());
    },
  };
}

export function useUpcoming() {
  let client = useClient();
  let loader = upcoming(client);
  loader.loadAndRefresh();
  return loader;
}
