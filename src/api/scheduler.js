import { computed, toRefs } from "@vue/composition-api";

import { makeState } from "./service-util.js";
import { useClient } from "./service.js";
import ScheduledArticle from "./scheduled-article.js";

export function useScheduler(id) {
  const { apiState, exec } = makeState();
  let client = useClient();

  const load = () => exec(() => client.article(id));
  load();

  return {
    ...toRefs(apiState),
    load,

    article: computed(() =>
      apiState.rawData
        ? new ScheduledArticle({ client, id, data: apiState.rawData })
        : null
    ),
  };
}
