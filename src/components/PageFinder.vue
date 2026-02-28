<script setup>
import { ref, computed, watch, onMounted } from "vue";
import draggable from "vuedraggable";

import { get, listPagesByFTS } from "@/api/client-v2.js";
import { Page } from "@/api/spotlightpa-page.js";

defineEmits(["select-page"]);

const query = ref("");
const rawPages = ref([]);
const error = ref(null);
const loading = ref(false);
let timeout = null;
let currentRequest = 0;

const pages = computed(() => rawPages.value.map((data) => new Page(data)));

async function fetchPages(q) {
  let token = ++currentRequest;
  const select = "-body";
  let [data, err] = await get(listPagesByFTS, { query: q, select });
  if (currentRequest !== token) {
    console.warn("discarding stale result");
    return;
  }
  loading.value = false;
  if (err) {
    error.value = err;
  } else {
    error.value = null;
    rawPages.value = data ?? [];
  }
}

watch(query, (val) => {
  loading.value = true;
  window.clearTimeout(timeout);
  timeout = window.setTimeout(() => {
    fetchPages(val);
  }, 300);
});

onMounted(() => {
  loading.value = true;
  fetchPages(query.value);
});
</script>

<template>
  <div>
    <BulmaFieldInput
      v-model="query"
      placeholder="Filter pages"
      label="Spotlight PA Pages"
    ></BulmaFieldInput>
    <draggable
      class="dropdown-content"
      :model-value="pages"
      item-key="id"
      :sort="false"
      :clone="(obj) => obj.filePath"
      :group="{ name: 'articles', pull: 'clone', put: false }"
      ghost-class="is-info"
      chosen-class="is-active"
    >
      <template #item="{ element: article }">
        <a
          class="dropdown-item select-none"
          @click="$emit('select-page', article)"
        >
          <span class="overflow">
            <b>{{ article.internalID }}</b
            >: {{ article.title }}
          </span>
        </a>
      </template>
      <template v-slot:footer>
        <div v-if="loading" class="dropdown-item">Loading…</div>
        <div v-else-if="rawPages.length === 0" class="dropdown-item">
          No results found
        </div>
        <div v-else-if="rawPages.length > 15" class="dropdown-item">
          More results hidden…
        </div>
      </template>
    </draggable>
    <ErrorSimple class="mt-4" :error="error"></ErrorSimple>
  </div>
</template>

<style scoped>
.overflow {
  text-overflow: ellipsis;
  overflow-x: hidden;
  display: block;
}
.select-none {
  cursor: grab;
  user-select: none;
}
</style>
