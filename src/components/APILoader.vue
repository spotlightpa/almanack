<script>
import { watch } from "@vue/composition-api";
import { useAPI } from "@/api/hooks.js";

export default {
  name: "APILoader",
  setup() {
    let { canLoadFeed, didLoad, error, isLoading, loadFeed } = useAPI();

    watch(() => {
      if (canLoadFeed.value && !didLoad.value) {
        loadFeed();
      }
    });

    return {
      canLoadFeed,
      error,
      isLoading,
      loadFeed,
    };
  },
};
</script>

<template>
  <div>
    <div v-if="!canLoadFeed" class="message is-warning">
      <p class="message-body">
        You don't have permission to view upcoming articles, sorry. Please
        contact
        <a href="mailto:cjohnson@spotlightpa.org">cjohnson@spotlightpa.org</a>
        to request access.
      </p>
    </div>
    <progress v-if="isLoading" class="progress is-large is-warning" max="100">
      Loading…
    </progress>
    <div v-if="error" class="message is-danger ">
      <div class="message-body">
        <p>{{ error }}</p>
        <div class="buttons">
          <button
            class="button is-danger has-text-weight-semibold"
            @click="loadFeed"
          >
            Reload?
          </button>
        </div>
      </div>
    </div>
    <div v-if="canLoadFeed && !isLoading">
      <slot></slot>
    </div>
  </div>
</template>
