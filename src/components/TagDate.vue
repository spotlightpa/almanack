<script setup>
import { formatDate } from "@/utils/time-format.js";

defineProps({ date: Date });
</script>

<template>
  <span class="tag is-light has-text-weight-semibold">
    {{ formatDate(date) }}
  </span>
</template>
