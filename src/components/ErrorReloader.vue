<script>
export default {
  props: {
    error: [Error, String],
  },
  emits: ["reload"],
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
    <div class="message-header">{{ name }}</div>
    <div class="message-body">
      <p class="content">{{ message }}</p>
      <div class="buttons">
        <button
          class="button is-danger has-text-weight-semibold"
          @click="$emit('reload')"
        >
          Reload?
        </button>
      </div>
    </div>
  </div>
</template>
