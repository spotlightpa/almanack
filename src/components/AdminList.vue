<script>
import { ref, computed } from "@vue/composition-api";

import AdminListPanel from "./AdminListPanel.vue";

export default {
  name: "AdminList",
  components: {
    AdminListPanel,
  },
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
    <div class="buttons has-addons">
      <button
        type="button"
        class="button has-text-weight-semibold"
        :class="show === 'a' ? 'is-primary is-selected' : 'is-light'"
        @click="show = 'a'"
      >
        Show all
      </button>
      <button
        type="button"
        class="button has-text-weight-semibold"
        :class="show === 'u' ? 'is-primary is-selected' : 'is-light'"
        @click="show = 'u'"
      >
        Unreleased
      </button>
      <button
        type="button"
        class="button has-text-weight-semibold"
        :class="show === 'r' ? 'is-primary is-selected' : 'is-light'"
        @click="show = 'r'"
      >
        Released
      </button>
    </div>

    <nav class="panel is-black">
      <h1 class="panel-heading" v-text="title"></h1>
      <article
        v-for="article of filteredArticles"
        :key="article.id"
        class="panel-block"
      >
        <AdminListPanel
          :article="article"
          @refresh="$emit('refresh', $event)"
        />
      </article>
    </nav>
  </div>
</template>
