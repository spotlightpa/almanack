import { computed, toRefs } from "@vue/composition-api";

import { makeState } from "./service-util.js";
import { useClient } from "./service.js";
import ScheduledArticle from "./scheduled-article.js";

function getScheduledArticle({ client, id }) {
  const { apiState, exec } = makeState();

  return {
    ...toRefs(apiState),

    article: computed(() =>
      apiState.rawData
        ? new ScheduledArticle({ client, id, data: apiState.rawData })
        : null
    ),

    load() {
      return exec(() => client.article(id));
    },
  };
}

export function useScheduler(id) {
  let obj = getScheduledArticle({ client: useClient(), id });
  obj.load();
  return obj;
}
