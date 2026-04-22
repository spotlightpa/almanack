<script setup>
import { watch, ref } from "vue";

import { watchAPI } from "@/api/service-util.js";
import { get, listPages } from "@/api/client-v2.js";

const props = defineProps({
  page: { default: "" },
});

const { apiState, fetch, computedProp } = watchAPI(
  () => props.page,
  (page) =>
    get(listPages, {
      page,
      path: "content/videos/",
      select: "frontmatter",
    })
);

const pages = ref([]);
watch(apiState.rawData, (data) => {
  if (data.pages) {
    pages.value = data.pages;
  }
});

const nextPage = computedProp("next_page", (page) => ({
  name: "video-pages",
  query: { page },
}));

function swap(event, i) {
  pages.value[i] = event;
}
</script>

<template>
  <MetaHead>
    <title>Videos • Spotlight PA Almanack</title>
  </MetaHead>

  <div>
    <BulmaBreadcrumbs
      :links="[
        { name: 'Admin', to: { name: 'admin' } },
        { name: 'Videos', to: {} },
      ]"
    ></BulmaBreadcrumbs>
    <h1 class="title">
      Videos
      <template v-if="page">(overflow page {{ page }})</template>
    </h1>

    <APILoader
      :is-loading="apiState.isLoading.value"
      :reload="fetch"
      :error="apiState.error.value"
    >
      <table class="table is-striped is-narrow is-fullwidth">
        <tbody>
          <tr v-for="(video, i) of pages" :key="video.id">
            <td>
              <VideoListRow
                :model-value="video"
                @update:model-value="swap($event, i)"
              />
            </td>
          </tr>
        </tbody>
      </table>

      <div class="buttons mt-5">
        <router-link
          v-if="nextPage"
          :to="nextPage"
          class="button is-primary has-text-weight-semibold"
        >
          Show More Videos…
        </router-link>
      </div>
    </APILoader>
  </div>
</template>
