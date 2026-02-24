<script setup>
import { computed } from "vue";

const props = defineProps({
  error: [Error, String],
});

const name = computed(() => props.error?.name ?? props.error ?? "Error");
const message = computed(() => props.error?.message ?? props.error ?? "");
</script>

<template>
  <div v-if="error" class="message is-danger">
    <p class="message-header">{{ name }}</p>
    <p class="message-body">{{ message }}</p>
  </div>
</template>
