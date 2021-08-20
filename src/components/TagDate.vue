<script>
import { formatDate } from "@/utils/time-format.js";

export default {
  name: "TagDate",
  props: { date: Date },
  setup() {
    return { formatDate };
  },
};
</script>

<template>
  <span class="tag is-light has-text-weight-semibold">
    {{ formatDate(date) }}
  </span>
</template>
