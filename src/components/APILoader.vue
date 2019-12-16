<script>
export default {
  name: "APILoader",
  props: {
    role: String,
  },
  created() {
    if (this.$auth.hasRole(this.role)) {
      this.$api.load();
    } else {
      this.$api.loading = false;
    }
  },
};
</script>

<template>
  <div>
    <div v-if="!$auth.hasRole(role)" class="message is-warning">
      <p class="message-body">
        You don't have permission to view upcoming articles, sorry. Please
        contact
        <a href="mailto:cjohnson@spotlightpa.org">cjohnson@spotlightpa.org</a>
        to request access.
      </p>
    </div>
    <progress
      v-if="$api.loadingRef.value"
      class="progress is-large is-warning"
      max="100"
    >
      Loading…
    </progress>
    <div v-if="$api.errorRef.value" class="message is-danger ">
      <div class="message-body">
        <p>{{ $api.error }}</p>
        <div class="buttons">
          <button
            class="button is-danger has-text-weight-semibold"
            @click="$api.reload"
          >
            Reload?
          </button>
        </div>
      </div>
    </div>
    <div v-if="$auth.hasRole(role) && !$api.loadingRef.value">
      <slot></slot>
    </div>
  </div>
</template>
