<script>
export default {
  props: {
    label: String,
  },
};
</script>

<template>
  <div>
    <p class="mt-1 label" v-text="label"></p>
    <nav class="mb-0 buttons">
      <slot />
    </nav>
  </div>
</template>
