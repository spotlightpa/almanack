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
    details() {
      return Object.entries(this.error?.details ?? {});
    },
  },
};
</script>

<template>
  <div v-if="error" class="message is-danger">
    <div class="message-header">{{ name }}</div>
    <div class="message-body">
      <p class="content">{{ message }}</p>
      <details>
        <p v-for="[key, vals] of details" :key="key" class="content">
          <strong v-text="key"></strong>:
          <span v-for="(val, i) of vals" :key="i" v-text="val"></span>
        </p>
      </details>
      <div class="mt-4 buttons">
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
