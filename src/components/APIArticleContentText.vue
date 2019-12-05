<script>
export default {
  props: {
    block: Object
  }
};
</script>

<template>
  <p v-html="block.content"></p>
</template>
