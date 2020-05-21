import { toRefs, computed } from "@vue/composition-api";

import ArcArticle from "./arc-article.js";
import { makeState } from "./service-util.js";
import { useClient } from "./service.js";

export function useAvailableList() {
  let { listAvailable } = useClient();
  let { apiState, exec } = makeState();

  const load = () => exec(listAvailable);
  load();

  return {
    ...toRefs(apiState),
    load,

    articles: computed(() =>
      apiState.isLoading || apiState.error || !apiState.rawData
        ? []
        : ArcArticle.from(apiState.rawData)
    ),
  };
}

export function getAvailableArticle(id) {
  let { getAvailable } = useClient();
  let { apiState, exec } = makeState();

  const load = () => exec(() => getAvailable(id));
  load();

  return {
    ...toRefs(apiState),
    load,

    article: computed(() =>
      apiState.isLoading || apiState.error || !apiState.rawData
        ? null
        : new ArcArticle(apiState.rawData)
    ),
  };
}

export function useUpcoming() {
  let { upcoming, listRefreshArc } = useClient();
  let { apiState, exec } = makeState();

  const actions = {
    load: () => exec(upcoming),
    loadAndRefresh: () => exec(listRefreshArc),
  };
  actions.loadAndRefresh();

  return {
    ...toRefs(apiState),
    ...actions,

    articles: computed(() =>
      apiState.isLoading || apiState.error || !apiState.rawData
        ? []
        : ArcArticle.from(apiState.rawData)
    ),
  };
}
