<script>
import { useScheduler } from "@/api/hooks.js";

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
    let { isLoading, load, error, article } = useScheduler(props.id);

    return {
      isLoading,
      load,
      error,
      article,
    };
  },
};
</script>

<template>
  <div>
    <nav class="breadcrumb has-succeeds-separator" aria-label="breadcrumbs">
      <ul>
        <li>
          <router-link :to="{ name: 'admin' }">Admin</router-link>
        </li>
        <li>
          <router-link exact :to="{ name: 'spotlightpa-articles' }">
            Spotlight PA Articles
          </router-link>

          (<router-link :to="article ? article.detailsRoute : ''"
            >Switch to Editor View</router-link
          >)
        </li>
      </ul>
    </nav>

    <APILoader :is-loading="isLoading" :reload="load" :error="error">
      <div v-if="!article" class="message is-warning">
        <p class="message-header">Not found</p>
        <p class="message-body">
          Article not found.
          <router-link :to="{ name: 'home' }">Go home</router-link>?
        </p>
      </div>
      <div v-else>
        <ScheduledArticleDetail :article="article" />
      </div>
    </APILoader>
  </div>
</template>
