<script>
import { reactive, toRefs } from "@vue/composition-api";
import { useClient } from "@/api/hooks.js";

import ArticleSlugLine from "./ArticleSlugLine.vue";

export default {
  name: "AdminListPanel",
  components: {
    ArticleSlugLine,
  },
  props: {
    article: Object,
  },
  setup() {
    let client = useClient();
    let apiStatus = reactive({
      error: null,
      isMakingPlanned: false,
      isMakingAvailable: false,
      isRemoving: false,
    });

    return {
      ...toRefs(apiStatus),

      async makePlanned(id) {
        apiStatus.isMakingPlanned = true;
        [, apiStatus.error] = await client.makePlanned(id);
        if (apiStatus.error) {
          apiStatus.isMakingPlanned = false;
          return;
        }
        let data;
        [data, apiStatus.error] = await client.upcoming();
        apiStatus.isMakingPlanned = false;
        if (apiStatus.error) {
          return;
        }
        this.$emit("refresh", data);
      },
      async makeAvailable(id) {
        apiStatus.isMakingAvailable = true;
        let data;
        [, apiStatus.error] = await client.makeAvailable(id);
        if (apiStatus.error) {
          apiStatus.isMakingAvailable = false;
          return;
        }
        [data, apiStatus.error] = await client.upcoming();
        apiStatus.isMakingAvailable = false;
        if (apiStatus.error) {
          return;
        }
        this.$emit("refresh", data);
      },
      async remove(id) {
        apiStatus.isRemoving = true;
        let data;
        [, apiStatus.error] = await client.removedArticle(id);
        if (apiStatus.error) {
          apiStatus.isRemoving = false;
          return;
        }
        [data, apiStatus.error] = await client.upcoming();
        apiStatus.isRemoving = false;
        if (apiStatus.error) {
          return;
        }
        this.$emit("refresh", data);
      },
    };
  },
};
</script>
<template>
  <div class="control">
    <h2 class="title is-spaced is-3">
      <ArticleSlugLine :article="article"></ArticleSlugLine>
    </h2>

    <p class="has-margin-top-negative">
      <strong>Byline:</strong>
      {{ article.authors | commaand }}
    </p>
    <p>
      <strong>Planned time:</strong>
      {{ article.plannedDate | formatDate }}
    </p>
    <div v-if="error">
      <p>{{ error }}</p>
    </div>
    <div class="buttons has-margin-top-thin">
      <button
        v-if="!article.isPlanned || article.isAvailable"
        type="button"
        class="button is-warning has-text-weight-semibold"
        :class="{ 'is-loading': isMakingPlanned }"
        @click="makePlanned(article.id)"
      >
        Show as Planned
      </button>
      <button
        v-if="!article.isAvailable"
        type="button"
        class="button is-success has-text-weight-semibold"
        :class="{ 'is-loading': isMakingAvailable }"
        @click="makeAvailable(article.id)"
      >
        Make Available
      </button>
      <button
        v-if="article.isPlanned || article.isAvailable"
        type="button"
        class="button is-light has-text-weight-semibold"
        :class="{ 'is-loading': isRemoving }"
        @click="remove(article.id)"
      >
        Remove from Almanack
      </button>
    </div>
  </div>
</template>
