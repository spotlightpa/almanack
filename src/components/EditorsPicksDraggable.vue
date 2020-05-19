<script>
import draggable from "vuedraggable";

export default {
  components: {
    draggable,
  },
  props: {
    value: Array,
  },
  methods: {
    remove(i) {
      let old = this.value;
      let newArray = old.slice(0, i).concat(old.slice(i + 1));
      this.$emit("input", newArray);
    },
    getComponentData() {
      return {
        attrs: {
          rows: "bulmaoverride",
          placeholder: "Drag text here",
        },
      };
    },
  },
};
</script>
<template>
  <draggable
    class="textarea"
    :list="value"
    :group="{ name: 'articles', pull: 'clone', put: true }"
    ghost-class="is-info"
    chosen-class="is-primary"
    :component-data="getComponentData()"
  >
    <span
      v-for="(article, i) of value"
      :key="i"
      class="tag is-medium spacer select-all"
    >
      {{ article.internal_id }}
      <button class="delete" @click="remove(i)"></button>
    </span>
    <span v-if="!value.length" slot="header" class="has-text-grey-lighter">
      Drag articles here
    </span>
  </draggable>
</template>

<style scoped>
.spacer {
  margin-right: 0.5rem;
  margin-bottom: 0.25rem;
}
.select-all {
  user-select: all;
}
</style>
