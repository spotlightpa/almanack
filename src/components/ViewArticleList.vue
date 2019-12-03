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
          :key="articleData.slug"
          class="panel-block"
        >
          <APIArticle v-slot="{ article }" :data="articleData">
            <div class="control">
              <p class="title is-spaced is-3">
                <router-link
                  :to="{ name: 'article', params: { id: article.id } }"
                >
                  <font-awesome-icon :icon="['far', 'newspaper']" />
                  {{ article.slug }}
                </router-link>
                <small v-if="article.isPublished" class="is-size-5">
                  (<a :href="article.pubURL" target="_blank"
                    >Live
                    <font-awesome-icon
                      :icon="['fas', 'external-link-alt']"
                    /> </a
                  >)
                </small>
              </p>
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
