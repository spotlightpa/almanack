import { reactive, computed, toRefs } from "@vue/composition-api";

import ScheduledArticle from "./scheduled-article.js";

export function getScheduledArticle({ client, id }) {
  const apiState = reactive({
    articleData: null,
    didLoad: false,
    isLoading: false,
    error: null,
    canLoad: client.hasAuthArticle(),
    article: computed(() =>
      apiState.articleData
        ? new ScheduledArticle({ client, id, data: apiState.articleData })
        : null
    ),
  });

  let methods = {
    async reload({ force = false } = {}) {
      if (apiState.isLoading && !force) {
        return;
      }
      apiState.didLoad = false;
      apiState.isLoading = true;
      [apiState.articleData, apiState.error] = await client.article(id);
      apiState.isLoading = false;
      if (apiState.articleData && apiState.error) {
        apiState.didLoad = true;
      }
    },
    async initLoad() {
      if (apiState.canLoad && !apiState.didLoad) {
        await methods.reload();
      }
    },
  };

  return {
    ...toRefs(apiState),
    ...methods,
  };
}
