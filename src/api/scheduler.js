import { computed, toRefs } from "@vue/composition-api";

import { makeState } from "./service-util.js";
import ScheduledArticle from "./scheduled-article.js";

export function getScheduledArticle({ client, id }) {
  const { apiState, exec } = makeState();

  return {
    ...toRefs(apiState),

    article: computed(() =>
      apiState.rawData
        ? new ScheduledArticle({ client, id, data: apiState.rawData })
        : null
    ),

    canLoad: client.hasAuthArticle(),
    load() {
      return exec(() => client.article(id));
    },
  };
}
