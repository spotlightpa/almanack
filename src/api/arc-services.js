import { toRefs, computed } from "@vue/composition-api";

import ArcArticle from "./arc-article.js";
import { makeState } from "./service-util.js";

export function listAvailable(client) {
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

export function getAvailable({ client, id }) {
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

export function upcoming(client) {
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
