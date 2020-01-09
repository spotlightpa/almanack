<script>
import { watch } from "@vue/composition-api";
import { useAuth, useAPI } from "@/api/hooks.js";

export default {
  name: "APILoader",
  props: {
    role: String,
  },
  setup(props) {
    let { hasRole } = useAuth();
    let { isLoading, error, loadFeed } = useAPI();

    let roleOK = hasRole(() => props.role);
    let didInitialLoad = false;
    watch(() => {
      if (roleOK.value && !didInitialLoad) {
        didInitialLoad = true;
        loadFeed();
      }
    });

    return {
      isLoading,
      loadFeed,
      error,
      roleOK,
    };
  },
};
</script>

<template>
  <div>
    <div v-if="!roleOK" class="message is-warning">
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
    <div v-if="roleOK && !isLoading">
      <slot></slot>
    </div>
  </div>
</template>
