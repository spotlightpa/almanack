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
    <router-link :to="article.detailsRoute">
      <font-awesome-icon :icon="['far', 'newspaper']" />
      {{ article.slug }}
    </router-link>

    <span class="is-inline-flex">
      <span class="tags has-addons">
        <span class="tag is-light">Status</span>
        <a
          v-if="article.isPublished"
          :href="article.pubURL"
          class="tag is-success"
          target="_blank"
          :title="`${article.slug} on SpotlightPA.org`"
        >
          <span class="icon is-size-6">
            <font-awesome-icon :icon="['fas', 'link']" />
          </span>
          {{ article.status | capfirst }}
        </a>
        <span v-else class="tag is-small" :class="tagStyle">{{
          article.status | capfirst
        }}</span>
        <a
          v-if="$auth.hasRole('arc user')"
          class="tag is-link"
          :href="article.arcURL"
          target="_blank"
        >
          <span class="icon is-size-6">
            <font-awesome-icon :icon="['fas', 'link']" />
          </span>
          <span>
            View in Arc
          </span>
        </a>
      </span>
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
