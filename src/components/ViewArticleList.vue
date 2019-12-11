<script>
import APIArticleSlugLine from "./APIArticleSlugLine.vue";
import APIArticleWordCount from "./APIArticleWordCount.vue";
import APILoader from "./APILoader.vue";

export default {
  name: "ViewArticleList",
  components: {
    APIArticleSlugLine,
    APIArticleWordCount,
    APILoader,
  },
};
</script>

<template>
  <div>
    <h2 class="title">
      Welcome, {{ $auth.fullName }}
      <small v-if="$auth.roles.length"> ({{ $auth.roles | commaand }}) </small>
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
    <APILoader role="editor">
      <nav class="panel is-black">
        <h1 class="panel-heading">
          Spotlight PA Articles
        </h1>
        <article
          v-for="article of $api.contents"
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
