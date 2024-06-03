<script>
export default {
  props: {
    label: String,
    icon: Array,
    size: {
      type: String,
      default: "is-small",
    },
    color: {
      type: String,
      default: "is-light",
    },
  },
};
</script>

<template>
  <button
    type="button"
    class="button has-text-weight-semibold"
    :class="[size, color]"
  >
    <span v-if="icon" class="icon">
      <font-awesome-icon :icon="icon"></font-awesome-icon>
    </span>
    <span v-text="label"></span>
  </button>
</template>
