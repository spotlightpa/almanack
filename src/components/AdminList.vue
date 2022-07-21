<script>
import { ref, computed } from "vue";

export default {
  name: "AdminList",
  props: {
    articles: Array,
    title: String,
  },
  setup(props) {
    let show = ref("a");
    return {
      show,
      filteredArticles: computed(() => {
        if (show.value === "a") {
          return props.articles;
        }
        if (show.value == "u") {
          return props.articles.filter((article) => !article.isAvailable);
        }
        return props.articles.filter((article) => article.isAvailable);
      }),
    };
  },
};
</script>
<template>
  <div>
    <nav class="panel is-black">
      <div class="panel-heading">
        <div class="level">
          <div class="level-left">
            <div class="level-item">
              <h1 class="has-text-weight-semibold">
                {{ title }}
              </h1>
            </div>
          </div>
          <div class="level-right">
            <div class="level-item">
              <div class="buttons has-addons">
                <button
                  type="button"
                  class="button is-small has-text-weight-semibold"
                  :class="show === 'a' ? 'is-primary is-selected' : 'is-light'"
                  @click="show = 'a'"
                >
                  Show all
                </button>
                <button
                  type="button"
                  class="button is-small has-text-weight-semibold"
                  :class="show === 'u' ? 'is-primary is-selected' : 'is-light'"
                  @click="show = 'u'"
                >
                  Unreleased
                </button>
                <button
                  type="button"
                  class="button is-small has-text-weight-semibold"
                  :class="show === 'r' ? 'is-primary is-selected' : 'is-light'"
                  @click="show = 'r'"
                >
                  Released
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
      <article
        v-for="article of filteredArticles"
        :key="article.id"
        class="panel-block pb-5 bg-striped"
      >
        <AdminListPanel
          :article="article"
          @refresh="$emit('refresh', $event)"
        />
      </article>
    </nav>
  </div>
</template>

<style scoped>
.bg-striped:nth-child(odd) {
  background-color: #eaeaea;
}
</style>
