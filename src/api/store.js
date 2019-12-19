import { reactive, computed, ref, toRefs } from "@vue/composition-api";
import Article from "./article.js";

export function makeAPI(service) {
  const feed = ref(null);

  const apiState = reactive({
    isLoading: true,
    error: null,
    articles: computed(() =>
      apiState.isLoading || apiState.error || !feed.value
        ? []
        : Article.from(feed.value)
    ),
  });

  let methods = {
    articleFromID(idFn) {
      return computed(() => {
        let id = idFn();
        return apiState.articles.find(article => article.id === id);
      });
    },
    async load() {
      if (!apiState.isLoading) {
        return;
      }
      [feed.value, apiState.error] = await service.upcoming();
      apiState.isLoading = false;
    },
    async reload() {
      apiState.isLoading = true;
      [feed.value, apiState.error] = await service.upcoming();
      apiState.isLoading = false;
    },
  };

  return {
    ...toRefs(apiState),
    ...methods,
  };
}
