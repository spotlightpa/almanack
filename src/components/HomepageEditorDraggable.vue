<script>
import draggable from "vuedraggable";

export default {
  components: {
    draggable,
  },
  props: {
    modelValue: Array,
  },
  methods: {
    remove(i) {
      let old = this.modelValue;
      let newArray = old.slice(0, i).concat(old.slice(i + 1));
      this.$emit("update:modelValue", newArray);
    },
    getComponentData() {
      return {
        attrs: {
          rows: "bulmaoverride",
        },
      };
    },
  },
};
</script>
<template>
  <draggable
    class="textarea"
    :list="modelValue"
    :group="{ name: 'articles', pull: 'clone', put: true }"
    ghost-class="is-info"
    chosen-class="is-primary"
    :component-data="getComponentData()"
  >
    <span
      v-for="(page, i) of modelValue"
      :key="i"
      class="tag is-medium spacer select-none"
    >
      {{ page.internalID }}
      <button class="delete" @click="remove(i)"></button>
    </span>
    <template v-slot:header>
      <span v-if="!modelValue.length" class="has-text-grey-lighter">
        Drag articles here
      </span></template
    >
  </draggable>
</template>

<style scoped>
.spacer {
  margin-right: 0.5rem;
  margin-bottom: 0.25rem;
}
.select-none {
  cursor: grab;
  user-select: none;
}
</style>
