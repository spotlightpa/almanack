<script>
export default {
  props: {
    block: Object,
  },
  computed: {
    listType() {
      return this.block.list_type === "unordered" ? "ul" : "ol";
    },
    items() {
      return this.block.items.filter((item) => item.type === "text");
    },
  },
};
</script>

<template>
  <div :is="listType">
    <li v-for="(item, i) of items" :key="i" v-html="item.content"></li>
  </div>
</template>
