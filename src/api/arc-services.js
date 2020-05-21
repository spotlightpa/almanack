import { toRefs, computed } from "@vue/composition-api";

import ArcArticle from "./arc-article.js";
import { makeState } from "./service-util.js";
import { useClient } from "./client.js";

export function useListAvailableArc() {
  let { listAvailableArc } = useClient();
  let { apiState, exec } = makeState();

  const load = () => exec(listAvailableArc);
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

export function useAvailableArc(id) {
  let { getAvailableArc } = useClient();
  let { apiState, exec } = makeState();

  const load = () => exec(() => getAvailableArc(id));
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
