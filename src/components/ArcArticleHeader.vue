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
  <component :is="hlevel" v-html="block.content"> </component>
</template>
