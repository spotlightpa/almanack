<script setup>
import { ref, computed } from "vue";

import { post, postPageLoad } from "@/api/client-v2.js";
import { makeState } from "@/api/service-util.js";

const { exec, apiStateRefs } = makeState();
const isLoading = apiStateRefs.isLoadingThrottled;
const { rawData, error } = apiStateRefs;
const pageId = computed(() => rawData.value);

const path = ref("");

function save() {
  return exec(() => post(postPageLoad, path.value));
}
</script>

<template>
  <MetaHead>
    <title>Page Load • Spotlight PA Almanack</title>
  </MetaHead>

  <div class="px-2">
    <BulmaBreadcrumbs
      :links="[
        { name: 'Admin', to: { name: 'admin' } },
        { name: 'Page Load', to: { name: 'page-load' } },
      ]"
    ></BulmaBreadcrumbs>
    <h1 class="title">Page Load</h1>
  </div>

  <label for="path" class="mt-4 label">Load into Almanack from Git</label>
  <form class="field is-grouped" @submit.prevent="save()">
    <div class="control is-expanded">
      <input
        id="path"
        v-model="path"
        class="input"
        placeholder="content/page.md"
      />
    </div>
    <div class="control">
      <BulmaPaste
        @paste="
          path = $event;
          save();
        "
      ></BulmaPaste>
    </div>
    <div class="control">
      <button
        class="button is-success has-text-weight-semibold"
        :class="isLoading && 'is-loading'"
        :disabled="!path || null"
      >
        Import
      </button>
    </div>
  </form>

  <p v-if="pageId">
    <RouterLink :to="{ name: 'news-page', params: { id: pageId } }">
      New page
    </RouterLink>
  </p>

  <ErrorSimple :error="error"></ErrorSimple>
</template>
