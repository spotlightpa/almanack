<script setup>
import { useAuth } from "@/api/auth.js";

defineProps({
  article: Object,
});

const { isSpotlightPAUser } = useAuth();
</script>

<template>
  <div
    v-if="isSpotlightPAUser && article.isGDoc && article.gdocs.warnings.length"
    class="message is-warning"
  >
    <div class="message-header">
      <span>
        <font-awesome-icon :icon="['fas', 'circle-exclamation']" />

        <span
          class="ml-1"
          v-text="article.gdocs.warnings.length === 1 ? 'Warning' : 'Warnings'"
        />
      </span>
    </div>

    <div class="message-body">
      <li v-for="(w, i) of article.gdocs.warnings" :key="i">
        <p v-text="w"></p>
      </li>
    </div>
  </div>
</template>
