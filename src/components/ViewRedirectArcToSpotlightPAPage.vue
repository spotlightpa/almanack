<script>
import { ref, toRefs, watch } from "@vue/composition-api";
import { makeState } from "@/api/service-util.js";
import { useClient } from "@/api/client.js";

export default {
  name: "ViewRedirectArcToSpotlightPAPage",
  components: {},
  props: {
    id: String,
  },
  metaInfo() {
    return {
      title: "Loading…",
    };
  },
  setup(props, { root }) {
    const isLoading = ref(false);
    window.setTimeout(() => {
      isLoading.value = true;
    }, 500);

    const { postPageForArcID } = useClient();
    const { apiState, exec } = makeState();

    const fetch = (id) =>
      exec(() => postPageForArcID({ "arc-id": id, "force-refresh": false }));

    watch(() => props.id, fetch, {
      immediate: true,
    });

    watch(
      () => apiState.rawData,
      (id) => {
        if (!id) {
          apiState.error = new Error("Could not load page");
          apiState.error.message = "No response from server";
          return;
        }
        root.$router.replace({
          name: "spotlightpa-page",
          params: {
            id: "" + id,
          },
        });
      }
    );

    const { error } = toRefs(apiState);

    return {
      fetch,
      error,
      isLoading,
    };
  },
};
</script>

<template>
  <div>
    <progress
      v-if="isLoading"
      class="my-5 progress is-large is-warning"
      max="100"
    >
      Loading…
    </progress>

    <div v-if="error" class="message is-danger">
      <div class="message-header">{{ error.name }}</div>
      <div class="message-body">
        <p class="content">{{ error.message }}</p>
        <div class="buttons">
          <button
            class="button is-danger has-text-weight-semibold"
            type="button"
            @click="fetch(id)"
          >
            Reload?
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
