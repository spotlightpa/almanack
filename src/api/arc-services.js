import { computed } from "@vue/composition-api";

import { loadItem, storeItem } from "@/utils/dom-utils.js";

import ArcArticle from "./arc-article.js";
import { useService } from "./service.js";

export function listAvailable(client) {
  let apiState = useService({
    canLoad: client.hasAuthAvailable(),
    serviceCall: () => client.listAvailable(),
  });

  return {
    ...apiState,
    articles: computed(() =>
      apiState.isLoading.value ||
      apiState.error.value ||
      !apiState.rawData.value
        ? []
        : ArcArticle.from(apiState.rawData.value)
    ),
  };
}

export function getAvailable({ client, id }) {
  let apiState = useService({
    canLoad: client.hasAuthAvailable(),
    serviceCall: () => client.getAvailable(id),
  });

  return {
    ...apiState,
    article: computed(() =>
      apiState.isLoading.value ||
      apiState.error.value ||
      !apiState.rawData.value
        ? null
        : new ArcArticle(apiState.rawData.value)
    ),
  };
}

const UPCOMING_KEY = "almanack:upcoming:cache";

export function upcoming(client) {
  let apiState = useService({
    canLoad: client.hasAuthUpcoming(),
    serviceCall: () => client.upcoming(),
  });

  return {
    ...apiState,
    articles: computed(() => {
      let rawItems;
      if (
        apiState.isLoading.value ||
        apiState.error.value ||
        !apiState.rawData.value
      ) {
        rawItems = loadItem(UPCOMING_KEY) ?? [];
      } else {
        rawItems = apiState.rawData.value;
        storeItem(UPCOMING_KEY, rawItems);
      }
      return ArcArticle.from(rawItems);
    }),
  };
}
