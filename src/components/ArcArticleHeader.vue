<script setup>
import { h } from "vue";

const props = defineProps({
  block: Object,
});
</script>

<template>
  <component :is="'h' + block.level" v-html="block.content"></component>
</template>
