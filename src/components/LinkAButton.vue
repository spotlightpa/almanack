<script>
export default {
  props: {
    href: String,
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
  <a
    :href="href"
    class="button has-text-weight-semibold"
    :class="[size, color]"
  >
    <span v-if="icon" class="icon">
      <font-awesome-icon :icon="icon" />
    </span>
    <span v-text="label" />
  </a>
</template>
