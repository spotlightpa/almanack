<script>
import { useAuth } from "@/api/auth.js";

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
      {{ article.internalID }}
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
        <TagLink
          v-if="isSpotlightPAUser && article.isGDoc"
          :to="article.adminRoute"
          :icon="['fas', 'sliders']"
        >
          Admin
        </TagLink>
        <TagLink
          v-if="isArcUser && article.isArc"
          :href="article.arc.arcURL"
          :icon="['fas', 'link']"
        >
          Arc
        </TagLink>
        <TagLink
          v-if="isSpotlightPAUser && article.isGDoc"
          :href="article.gdocsURL"
          :icon="['fas', 'link']"
        >
          Google Docs
        </TagLink>
        <TagLink
          v-if="isSpotlightPAUser && article.pageRoute"
          :to="article.pageRoute"
          :icon="['fas', 'user-clock']"
        >
          Spotlight admin
        </TagLink>
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
