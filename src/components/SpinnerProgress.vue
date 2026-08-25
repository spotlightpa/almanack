<script setup>
const props = defineProps({
  isLoading: Boolean,
});
</script>

<template>
  <progress
    v-if="isLoading"
    class="my-5 progress is-large is-warning"
    max="100"
  >
    Loading…
  </progress>
</template>
