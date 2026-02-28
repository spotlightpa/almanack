<script setup>
defineProps({
  block: Object,
});
</script>

<template>
  <!-- eslint-disable-next-line vue/no-v-text-v-html-on-component -->
  <component :is="'h' + block.level" v-html="block.content"></component>
</template>
