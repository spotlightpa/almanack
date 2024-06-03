<script setup>
import { ref } from "vue";
import { useRouter } from "vue-router";

import { get, getSharedArticle } from "@/api/client-v2.js";

const router = useRouter();
const props = defineProps({
  id: String,
  sourceType: {
    default: "arc",
  },
});

const isLoading = ref(false);
const isLoadingDebounced = ref(false);
window.setTimeout(() => {
  isLoadingDebounced.value = true;
}, 500);
const error = ref(null);

async function load() {
  isLoading.value = true;
  let [article, err] = await get(getSharedArticle, {
    source_type: props.sourceType,
    source_id: props.id,
  });
  isLoading.value = false;
  if (err) {
    error.value = err;
    return;
  }
  router.replace({
    name: "shared-article",
    params: {
      id: "" + article.id,
    },
  });
}

load();
</script>

<template>
  <SpinnerProgress
    :is-loading="isLoading && isLoadingDebounced"
  ></SpinnerProgress>
  <ErrorReloader :error="error" @reload="load"></ErrorReloader>
</template>
