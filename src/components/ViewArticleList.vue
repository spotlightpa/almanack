<script>
import ArticleList from "./ArticleList.vue";
import APILoader from "./APILoader.vue";
import { useAuth, useFeed } from "@/api/hooks.js";

export default {
  name: "ViewArticleList",
  components: {
    ArticleList,
    APILoader,
  },
  setup() {
    let { fullName, roles, isSpotlightPAUser } = useAuth();
    let { articles, canLoad, isLoading, fetch, error } = useFeed();

    return {
      canLoad,
      isLoading,
      fetch,
      error,
      fullName,
      roles,
      articles,
      isSpotlightPAUser,
    };
  },
};
</script>

<template>
  <div>
    <h2 class="title">
      Welcome, {{ fullName }}
      <small v-if="roles.length > 1"> ({{ roles | commaand }}) </small>
      <small v-if="roles.length === 0"> (Not Authorized) </small>
    </h2>
    <p class="content">
      Please note that this is an internal content distribution system, not
      intended for public use. Please
      <strong>do not share this URL</strong> with anyone besides the appointed
      contacts at your organization and please be mindful of the notes and
      embargos attached to each story. For assistance or if you have any
      questions, please contact Sarah Anne Hughes (<a
        href="mailto:shughes@spotlightpa.org"
        >shughes@spotlightpa.org</a
      >).
    </p>

    <div v-if="isSpotlightPAUser" class="block">
      <router-link
        :to="{ name: 'uploader' }"
        class="button is-success has-text-weight-semibold"
      >
        <span class="icon">
          <font-awesome-icon :icon="['fa', 'file-upload']" />
        </span>
        <span>
          Upload images
        </span>
      </router-link>
    </div>

    <APILoader
      :can-load="canLoad"
      :is-loading="isLoading"
      :reload="fetch"
      :error="error"
    >
      <ArticleList
        v-if="articles.length"
        :articles="articles"
        title="Spotlight PA Articles"
      ></ArticleList>
    </APILoader>
  </div>
</template>
