<script>
export default {
  props: {
    article: {
      type: Object,
      required: true,
    },
  },
  computed: {
    tagStyle() {
      return {
        published: "is-success",
        ready: "is-warning",
        "not ready": "is-danger",
      }[this.article.status];
    },
  },
};
</script>

<template>
  <span>
    <router-link
      :to="{ name: 'article', params: { id: article.id } }"
      class="middle"
    >
      <font-awesome-icon :icon="['far', 'newspaper']" />
      {{ article.slug }}
    </router-link>

    <span class="tags is-inline-flex has-addons middle">
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
      <span v-else class="tag is-small" :class="tagStyle">{{
        article.status | capfirst
      }}</span>
    </span>
  </span>
</template>

<style scoped>
.router-link-exact-active {
  pointer-events: none;
  color: inherit;
}

.middle {
  vertical-align: middle;
}
</style>
