<script setup>
import { ref, watch } from "vue";
import { debounce as makeDebounce } from "@/utils/wait.ts";

const props = defineProps({
  isLoading: Boolean,
  debounce: Number,
});

const debouncedLoading = ref(true);
const setLoadingTrue = makeDebounce(props.debounce ?? 0, () => {
  debouncedLoading.value = true;
});
watch(
  () => props.isLoading,
  (val) => {
    if (!props.debounce || !val) return;
    debouncedLoading.value = false;
    setLoadingTrue();
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
