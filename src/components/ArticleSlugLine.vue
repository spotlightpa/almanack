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
    let { isArcUser, isSpotlightPAUser } = useAuth();

    return {
      isSpotlightPAUser,
      isArcUser,
    };
  },
  computed: {
    tagStyle() {
      if (this.article.status === "released") {
        return "is-success";
      }
      if (this.article.status === "imported") {
        return "is-danger";
      }
      return "is-warning";
    },
  },
};
</script>

<template>
  <span>
    <router-link :to="article.detailsRoute" class="mr-2 middle">
      <font-awesome-icon :icon="['far', 'newspaper']" />
      {{ article.slug }}
    </router-link>
    <span class="is-inline-flex middle">
      <span class="tags">
        <span class="tag is-small" :class="tagStyle">
          <span class="icon is-size-6">
            <font-awesome-icon
              :icon="
                article.isShared ? ['fas', 'check-circle'] : ['fas', 'pen-nib']
              "
            />
          </span>
          <span v-text="article.statusVerbose"></span>
        </span>
        <a
          v-if="isArcUser && article.isArc"
          class="tag is-light"
          :href="article.arc.arcURL"
          target="_blank"
        >
          <span class="icon is-size-6">
            <font-awesome-icon :icon="['fas', 'link']" />
          </span>
          <span>Arc view</span>
        </a>
        <router-link
          v-if="isSpotlightPAUser && article.pageRoute"
          class="tag is-light"
          :to="article.pageRoute"
        >
          <span class="icon">
            <font-awesome-icon :icon="['fas', 'user-clock']" />
          </span>
          <span>Spotlight PA page</span>
        </router-link>
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
