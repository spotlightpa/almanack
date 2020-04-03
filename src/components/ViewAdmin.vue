<script>
import { ref } from "@vue/composition-api";

import AdminList from "./AdminList.vue";
import APILoader from "./APILoader.vue";
import EmailComposer from "./EmailComposer.vue";

import { useUpcoming } from "@/api/hooks.js";

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
    let { articles, rawData, canLoad, isLoading, fetch, error } = useUpcoming();

    return {
      showComposer: ref(false),

      canLoad,
      isLoading,
      fetch,
      error,
      articles,

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
    <p class="content">
      Tools:
      <a
        href="javascript:
(()=>{
  let match = location.href.match(/\/\d{4}\/\d\d\/([\w-]+)\/?$/);
  if (!match) {
    alert('Not on Spotlight PA article');
    return;
  }
  let [, slug] = match;
  window.location = 'https://almanack.data.spotlightpa.org/api/bookmarklet/' + slug;
})();
        "
      >
        Jump to admin
      </a>
      bookmarklet
      <sup
        ><a
          title="What's a bookmarklet?"
          href="https://support.mozilla.org/en-US/kb/bookmarklets-perform-common-web-page-tasks"
          >?</a
        ></sup
      >
    </p>

    <div class="level">
      <div class="level-left">
        <div class="level-item">
          <div class="control">
            <label class="label">Upload an image</label>
            <ImageUploader />
          </div>
        </div>
        <div class="level-item">
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
              <span
                v-text="!showComposer ? 'Compose Message' : 'Hide Message'"
              />
            </button>
          </div>
        </div>
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

    <div class="buttons">
      <router-link
        :to="{ name: 'articles' }"
        class="button is-small is-success has-text-weight-semibold"
      >
        <span class="icon">
          <font-awesome-icon :icon="['far', 'newspaper']" />
        </span>
        <span>
          Editor Article List
        </span>
      </router-link>

      <router-link
        class="button is-small is-success has-text-weight-semibold"
        :to="{ name: 'spotlightpa-articles' }"
      >
        <span class="icon">
          <font-awesome-icon :icon="['far', 'newspaper']" />
        </span>
        <span>
          Spotlight PA Articles
        </span>
      </router-link>

      <router-link
        :to="{ name: 'uploader' }"
        class="button is-small is-success has-text-weight-semibold"
      >
        <span class="icon">
          <font-awesome-icon :icon="['fa', 'file-upload']" />
        </span>
        <span>
          Images
        </span>
      </router-link>

      <router-link
        :to="{ name: 'domains' }"
        class="button is-small is-success has-text-weight-semibold"
      >
        <span class="icon">
          <font-awesome-icon :icon="['fas', 'user-circle']" />
        </span>
        <span>
          Approved User Domains
        </span>
      </router-link>
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
