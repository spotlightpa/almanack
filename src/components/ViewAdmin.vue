<script>
import AdminList from "./AdminList.vue";
import APILoader from "./APILoader.vue";
import { useAuth, useUpcoming } from "@/api/hooks.js";

import ImageUploader from "./ImageUploader.vue";

export default {
  name: "ViewAdmin",
  components: {
    AdminList,
    APILoader,
    ImageUploader,
  },
  setup() {
    let { fullName, roles, isSpotlightPAUser } = useAuth();
    let { articles, rawData, canLoad, isLoading, fetch, error } = useUpcoming();

    return {
      canLoad,
      isLoading,
      fetch,
      error,
      fullName,
      roles,
      articles,
      isSpotlightPAUser,
      refresh(newData) {
        rawData.value = newData;
      },
    };
  },
};
</script>

<template>
  <div>
    <h1 class="title">
      Spotlight Administrator
    </h1>

    <div v-if="isSpotlightPAUser" class="block">
      <h2 class="subtitle is-4 has-text-weight-semibold">Upload an image</h2>
      <ImageUploader />
    </div>

    <APILoader
      :can-load="canLoad"
      :is-loading="isLoading"
      :reload="fetch"
      :error="error"
    >
      <keep-alive>
        <AdminList
          v-if="articles.length"
          :articles="articles"
          title="Arc Articles"
          @refresh="refresh"
        />
      </keep-alive>
    </APILoader>
  </div>
</template>
