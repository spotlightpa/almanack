<script>
import APIArticleSlugLine from "./APIArticleSlugLine.vue";
import APILoader from "./APILoader.vue";

export default {
  name: "ViewArticleList",
  components: {
    APIArticleSlugLine,
    APILoader,
  },
};
</script>

<template>
  <div>
    <h2 class="title">
      Welcome, {{ $auth.fullName }}
      <small v-if="$auth.roles.length">
        (<span v-for="role of $auth.roles" :key="role" v-text="role"></span>)
      </small>
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
              <strong>Planned for:</strong>
              {{ article.plannedDate | formatDate }}
            </p>
            <p class="has-margin-top-thin content is-small">
              {{ article.budgetLine }}
            </p>

            <div class="level is-mobile is-clipped">
              <div class="level-left">
                <p v-if="article.plannedWordCount" class="level-item">
                  <span>
                    <strong>Planned Word Count:</strong>
                    {{ article.plannedWordCount | intcomma }}
                  </span>
                </p>

                <p
                  v-if="article.actualWordCount"
                  class="level-item is-hidden-mobile"
                >
                  <span>
                    <strong>Word Count:</strong>
                    {{ article.actualWordCount | intcomma }}
                  </span>
                </p>
                <p class="level-item is-hidden-mobile">
                  <span>
                    <strong>Lines:</strong>
                    {{ article.actualLineCount }}
                  </span>
                </p>
                <p class="level-item is-hidden-mobile">
                  <span>
                    <strong>Column inches:</strong>
                    {{ article.actualInchCount }}
                  </span>
                </p>
              </div>
            </div>
          </div>
        </article>
      </nav>
    </APILoader>
  </div>
</template>
