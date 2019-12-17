<script>
import { useAuth } from "@/api/hooks.js";

export default {
  props: {
    article: {
      type: Object,
      required: true,
    },
  },
  setup() {
    let { hasRole } = useAuth();

    return {
      isSpotlightPAUser: hasRole("Spotlight PA"),
      isArcUser: hasRole("arc user"),
    };
  },
  computed: {
    tagStyle() {
      return {
        published: "is-success",
        ready: "is-warning",
        "not ready": "is-danger",
      }[this.article.status];
    },
    linkTag() {
      if (this.isSpotlightPAUser) {
        return "router-link";
      }
      return this.article.status !== "not ready" ? "router-link" : "span";
    },
  },
};
</script>

<template>
  <span>
    <router-link :is="linkTag" :to="article.detailsRoute" class="middle">
      <font-awesome-icon :icon="['far', 'newspaper']" />
      {{ article.slug }}
    </router-link>

    <span class="is-inline-flex middle">
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
          <span>
            {{ article.status | capfirst }}
          </span>
        </a>
        <span v-else class="tag is-small" :class="tagStyle">{{
          article.status | capfirst
        }}</span>
        <a
          v-if="isArcUser"
          class="tag is-primary"
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
