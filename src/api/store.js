import { ref, computed } from "@vue/composition-api";
import Article from "./article.js";
import { useService } from "./service.js";

function makeAPI() {
  let service = useService();

  const loadingRef = ref(true);
  const feedRef = ref(null);
  const errorRef = ref(null);

  let contents = computed(() =>
    loadingRef.value || errorRef.value ? [] : Article.from(feedRef.value)
  );

  return {
    loadingRef,
    errorRef,

    contents,

    getByID(id) {
      return contents.value.find(article => article.id === id);
    },
    async load() {
      if (!loadingRef.value) {
        return;
      }
      [feedRef.value, errorRef.value] = await service.upcoming();
      loadingRef.value = false;
    },
    async reload() {
      loadingRef.value = true;
      [feedRef.value, errorRef.value] = await service.upcoming();
      loadingRef.value = false;
    },
  };
}

let $api;

export function useAPI() {
  if (!$api) {
    $api = makeAPI();
  }
  return $api;
}
