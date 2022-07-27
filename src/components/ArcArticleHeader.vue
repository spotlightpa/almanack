<script>
export default {
  props: {
    block: Object,
  },
  computed: {
    hlevel() {
      return `h${this.block.level}`;
    },
  },
};
</script>

<template>
  <!-- eslint-disable  vue/no-v-text-v-html-on-component-->
  <component :is="hlevel" v-html="block.content" />
  <!-- eslint-enable -->
</template>
