<script>
export default {
  name: "APILoader",
  props: {
    isLoading: Boolean,
    reload: Function,
    error: Error,
  },
};
</script>

<template>
  <div>
    <progress v-if="isLoading" class="progress is-large is-warning" max="100">
      Loading…
    </progress>
    <div v-if="error" class="message is-danger">
      <div class="message-header">{{ error.name }}</div>
      <div class="message-body">
        <p class="content">{{ error.message }}</p>
        <div class="buttons">
          <button
            class="button is-danger has-text-weight-semibold"
            @click="reload"
          >
            Reload?
          </button>
        </div>
      </div>
    </div>
    <div v-if="!isLoading && !error">
      <slot></slot>
    </div>
  </div>
</template>
