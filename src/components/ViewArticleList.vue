<script>
import APIArticleSlugLine from "./APIArticleSlugLine.vue";
import APIArticleWordCount from "./APIArticleWordCount.vue";
import APILoader from "./APILoader.vue";
import { useAuth, useAPI } from "@/api/hooks.js";

export default {
  name: "ViewArticleList",
  components: {
    APIArticleSlugLine,
    APIArticleWordCount,
    APILoader,
  },
  setup() {
    let { fullName, roles } = useAuth();
    let { articles, canLoad, isLoading, initLoad, reload, error } = useAPI();

    initLoad();

    return {
      canLoad,
      isLoading,
      reload,
      error,
      fullName,
      roles,
      articles,
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
      questions, please contact Joanna Bernstein (<a
        href="mailto:joanna@spotlightpa.org"
        >joanna@spotlightpa.org</a
      >).
    </p>
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
