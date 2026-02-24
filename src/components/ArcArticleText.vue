<script setup>
defineProps({
  block: Object,
});
</script>

<template>
  <p v-html="block.content"></p>
</template>
