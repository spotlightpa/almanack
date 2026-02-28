<script setup>
import { computed } from "vue";

const props = defineProps({
  block: Object,
});

const listType = computed(() =>
  props.block.list_type === "unordered" ? "ul" : "ol"
);
const items = computed(() =>
  props.block.items.filter((item) => item.type === "text")
);
</script>

<template>
  <component :is="listType">
    <li v-for="(item, i) of items" :key="i" v-html="item.content"></li>
  </component>
</template>
