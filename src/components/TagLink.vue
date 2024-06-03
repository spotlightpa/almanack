<script setup>
defineProps({
  href: { type: String, required: false },
  to: { required: false },
  icon: { required: false },
});
</script>

<template>
  <a
    v-if="href"
    class="tag is-light has-text-weight-semibold"
    :href="href"
    target="_blank"
  >
    <span v-if="icon" class="icon is-size-6">
      <font-awesome-icon :icon="icon"></font-awesome-icon>
    </span>
    <span><slot></slot></span>
  </a>
  <router-link
    v-else-if="to"
    class="tag is-light has-text-weight-semibold"
    :to="to"
  >
    <span v-if="icon" class="icon is-size-6">
      <font-awesome-icon :icon="icon"></font-awesome-icon>
    </span>
    <span><slot></slot></span>
  </router-link>
  <span v-else class="tag is-light has-text-weight-semibold">
    <span v-if="icon" class="icon is-size-6">
      <font-awesome-icon :icon="icon"></font-awesome-icon>
    </span>
    <span><slot></slot></span>
  </span>
</template>
