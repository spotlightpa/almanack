import { computed } from "@vue/composition-api";
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

export function available({ client, id }) {
  let apiState = useService({
    canLoad: client.hasAuthAvailable(),
    serviceCall: () => client.available(id),
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

export function upcoming(client) {
  let apiState = useService({
    canLoad: client.hasAuthUpcoming(),
    serviceCall: () => client.upcoming(),
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
