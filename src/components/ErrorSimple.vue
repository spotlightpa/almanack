<script>
export default {
  props: {
    error: [Error, String],
  },
  computed: {
    name() {
      return this.error?.name ?? this.error ?? "Error";
    },
    message() {
      return this.error?.message ?? this.error ?? "";
    },
  },
};
</script>

<template>
  <div v-if="error" class="message is-danger">
    <p class="message-header">{{ name }}</p>
    <p class="message-body">{{ message }}</p>
  </div>
</template>
