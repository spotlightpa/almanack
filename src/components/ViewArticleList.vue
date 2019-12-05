<script>
import APIArticle from "./APIArticle.vue";
import APILoader from "./APILoader.vue";

export default {
  name: "ViewArticleList",
  components: {
    APIArticle,
    APILoader
  }
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
    <APILoader role="editor">
      <nav class="panel is-black">
        <h1 class="panel-heading">
          Spotlight PA Articles
        </h1>
        <article
          v-for="articleData of $api.contents"
          :key="articleData._id"
          class="panel-block"
        >
          <APIArticle v-slot="{ article }" :data="articleData">
            <div class="control">
              <h2 class="title is-spaced is-3">
                <router-link
                  :to="{ name: 'article', params: { id: article.id } }"
                >
                  <font-awesome-icon :icon="['far', 'newspaper']" />
                  {{ article.slug }}
                </router-link>

                <span class="tags is-inline-flex has-addons">
                  <span class="tag is-light">Status</span>
                  <a
                    v-if="article.isPublished"
                    :href="article.pubURL"
                    class="tag is-success"
                    target="_blank"
                    :title="`${article.slug} on SpotlightPA.org`"
                  >
                    <span class="is-size-6">
                      <font-awesome-icon class="" :icon="['fas', 'link']" />
                    </span>
                    {{ article.status | capfirst }}
                  </a>
                  <span v-else class="tag is-warning is-small">{{
                    article.status | capfirst
                  }}</span>
                </span>
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
          </APIArticle>
        </article>
      </nav>
    </APILoader>
  </div>
</template>
