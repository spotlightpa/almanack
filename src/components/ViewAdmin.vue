<script>
import { ref } from "@vue/composition-api";

import AdminList from "./AdminList.vue";
import APILoader from "./APILoader.vue";
import EmailComposer from "./EmailComposer.vue";

import { useAuth, useUpcoming } from "@/api/hooks.js";

import ImageUploader from "./ImageUploader.vue";

export default {
  name: "ViewAdmin",
  components: {
    AdminList,
    APILoader,
    EmailComposer,
    ImageUploader,
  },
  metaInfo: {
    title: "Admin",
  },
  setup() {
    let { fullName, roles, isSpotlightPAUser } = useAuth();
    let { articles, rawData, canLoad, isLoading, fetch, error } = useUpcoming();

    return {
      showComposer: ref(false),

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

    <div v-if="isSpotlightPAUser" class="field is-grouped">
      <div class="control">
        <label class="label">Upload an image</label>
        <ImageUploader />
      </div>
      <div class="control">
        <label class="label">Compose a message</label>
        <button
          type="button"
          class="button is-primary has-text-weight-semibold"
          @click="showComposer = !showComposer"
        >
          <span class="icon">
            <font-awesome-icon :icon="['fas', 'paper-plane']" />
          </span>
          <span v-text="!showComposer ? 'Compose Message' : 'Hide Message'" />
        </button>
      </div>
    </div>

    <keep-alive>
      <EmailComposer
        v-if="showComposer"
        initial-subject="Subject"
        initial-body="Email body"
        @hide="showComposer = false"
      />
    </keep-alive>

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
