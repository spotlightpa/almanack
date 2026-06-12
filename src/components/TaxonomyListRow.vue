<script setup>
import { computed } from "vue";

import { TaxonomyPage } from "@/api/taxonomy-page.js";

const props = defineProps({
  modelValue: {
    type: Object,
    required: true,
  },
});

const page = computed(() => new TaxonomyPage(props.modelValue));
</script>

<template>
  <router-link
    class="is-flex-tablet my-2 is-align-items-center"
    :to="page.taxoPage"
  >
    <div class="is-flex-grow-1">
      <p class="mt-0 has-text-weight-semibold has-text-black">
        {{ page.title }}
      </p>
      <p class="has-text-weight-light has-text-dark">
        {{ page.linkTitle }}
      </p>
    </div>
    <div
      v-if="page.image"
      class="m-2 is-flex-shrink-0 is-clipped"
      style="width: 128px"
    >
      <picture class="image has-ratio">
        <img
          class="is-3x4"
          :src="page.thumbnailURL"
          :alt="page.imageDescription"
          loading="lazy"
        />
      </picture>
    </div>
  </router-link>
</template>
