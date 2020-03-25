<script>
import { useScheduler, useAuth } from "@/api/hooks.js";

import APILoader from "./APILoader.vue";
import ScheduledArticleDetail from "./ScheduledArticleDetail.vue";

export default {
  name: "ViewArticleSchedule",
  components: {
    APILoader,
    ScheduledArticleDetail,
  },
  props: {
    id: String,
  },
  metaInfo() {
    return {
      title: this.article ? `Schedule ${this.article.id}` : "Schedule Article",
    };
  },
  setup(props) {
    let { isSpotlightPAUser } = useAuth();
    let { canLoad, isLoading, reload, error, article } = useScheduler(props.id);

    return {
      isSpotlightPAUser,
      canLoad,
      isLoading,
      reload,
      error,
      article,
    };
  },
};
</script>

<template>
  <div>
    <div v-if="!isSpotlightPAUser" class="message is-danger">
      <p class="message-header">Not Authorized</p>

      <p class="message-body">
        You do not have permission to use this page.
        <strong
          ><router-link :to="{ name: 'home' }">Go home</router-link>?</strong
        >
      </p>
    </div>
    <APILoader
      v-else
      :can-load="canLoad"
      :is-loading="isLoading"
      :reload="reload"
      :error="error"
    >
      <div v-if="!article" class="message is-warning">
        <p class="message-header">
          Not found
        </p>
        <p class="message-body">
          Article not found.
          <router-link :to="{ name: 'home' }">Go home</router-link>?
        </p>
      </div>
      <div v-else>
        <ScheduledArticleDetail :article="article"></ScheduledArticleDetail>
      </div>
    </APILoader>
  </div>
</template>
