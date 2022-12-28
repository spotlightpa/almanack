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
        <span class="tag is-small" :class="article.statusClass">
          <span class="icon is-size-6">
            <font-awesome-icon
              :icon="
                article.isShared ? ['fas', 'check-circle'] : ['fas', 'pen-nib']
              "
            />
          </span>
          <span v-text="article.statusVerbose"></span>
        </span>
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
