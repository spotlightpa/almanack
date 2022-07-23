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
  <component :is="hlevel">
    <template v-html="block.content" />
  </component>
</template>
