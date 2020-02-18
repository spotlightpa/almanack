<script>
import APIArticleSlugLine from "./APIArticleSlugLine.vue";
import APIArticleWordCount from "./APIArticleWordCount.vue";
import APILoader from "./APILoader.vue";
import { useAuth, useFeed } from "@/api/hooks.js";

export default {
  name: "ViewArticleList",
  components: {
    APIArticleSlugLine,
    APIArticleWordCount,
    APILoader,
  },
  setup() {
    let { fullName, roles, isSpotlightPAUser } = useAuth();
    let { articles, canLoad, isLoading, reload, error } = useFeed();

    return {
      canLoad,
      isLoading,
      reload,
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
      <small v-if="roles.length"> ({{ roles | commaand }}) </small>
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
      :reload="reload"
      :error="error"
    >
      <nav class="panel is-black">
        <h1 class="panel-heading">
          Spotlight PA Articles
        </h1>
        <article
          v-for="article of articles"
          :key="article.id"
          class="panel-block"
        >
          <div class="control">
            <h2 class="title is-spaced is-3">
              <APIArticleSlugLine :article="article"></APIArticleSlugLine>
            </h2>

            <p class="has-margin-top-negative">
              <strong>Byline:</strong>
              {{ article.authors | commaand }}
            </p>
            <p>
              <strong>Planned time:</strong>
              {{ article.plannedDate | formatDate }}
            </p>
            <p class="has-margin-top-thin content is-small">
              {{ article.budgetLine }}
            </p>

            <APIArticleWordCount :article="article"></APIArticleWordCount>
          </div>
        </article>
      </nav>
    </APILoader>
  </div>
</template>
