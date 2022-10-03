<script>
export default {
  props: {
    value: {
      type: String,
      required: true,
    },
    warn: {
      type: Number,
      default: 15,
    },
    max: {
      type: Number,
      default: 20,
    },
  },
};
</script>

<template>
  <div v-if="value.length > warn" class="is-flex">
    <span class="has-text-weight-semibold"> Characters used: </span>
    <span class="ml-3"> {{ value.length }} / {{ max }} </span>
    <progress
      class="mt-1 ml-3 progress max-w"
      :value="value.length"
      :max="max"
      :class="value.length < max ? 'is-warning' : 'is-danger'"
    ></progress>
  </div>
</template>

<style scoped>
.max-w {
  max-width: 200px;
}
</style>
