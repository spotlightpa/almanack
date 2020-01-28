<script>
export default {
  name: "APILoader",
  props: {
    canLoad: Boolean,
    isLoading: Boolean,
    reload: Function,
    error: Error,
  },
};
</script>

<template>
  <div>
    <div v-if="!canLoad" class="message is-warning">
      <p class="message-body">
        You don't have permission to view this page. Please contact
        <a href="mailto:cjohnson@spotlightpa.org">cjohnson@spotlightpa.org</a>
        to request access.
      </p>
    </div>
    <progress v-if="isLoading" class="progress is-large is-warning" max="100">
      Loading…
    </progress>
    <div v-if="error" class="message is-danger ">
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
    <div v-if="canLoad && !isLoading">
      <slot></slot>
    </div>
  </div>
</template>
