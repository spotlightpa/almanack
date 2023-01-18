<script setup>
import { ref, watch } from "vue";

const props = defineProps({
  isLoading: Boolean,
  debounce: Number,
});

const debouncedLoading = ref(true);
let timeout = null;
watch(
  () => props.isLoading,
  (val) => {
    if (!props.debounce || !val) return;
    window.clearTimeout(timeout);
    debouncedLoading.value = false;
    timeout = window.setTimeout(() => {
      debouncedLoading.value = true;
    }, props.debounce);
  }
);
</script>

<template>
  <progress
    v-if="isLoading && debouncedLoading"
    class="my-5 progress is-large is-warning"
    max="100"
  >
    Loading…
  </progress>
</template>
