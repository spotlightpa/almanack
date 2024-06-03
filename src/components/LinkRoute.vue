<script>
export default {
  props: {
    to: String,
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
  <router-link
    :to="{ name: to }"
    class="button has-text-weight-semibold"
    :class="[size, color]"
  >
    <span v-if="icon" class="icon">
      <font-awesome-icon :icon="icon"></font-awesome-icon>
    </span>
    <span v-text="label"></span>
  </router-link>
</template>
