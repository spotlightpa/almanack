<script>
import { defineComponent, ref, watch } from "vue";

import { makeState } from "@/api/service-util.js";
import { useClient } from "@/api/client.js";

export default defineComponent({
  setup(_, { root }) {
    const url = ref("");
    const { getPageByURLPath } = useClient();
    const { apiStateRefs, exec } = makeState();

    async function lookup() {
      const select = "-body";
      let path = url.value;
      try {
        path = new URL(path).pathname;
        // eslint-disable-next-line no-empty
      } catch (e) {}
      await exec(() => getPageByURLPath({ params: { path, select } }));
    }
    watch(apiStateRefs.rawData, (page) => {
      if (!page) {
        return;
      }
      let name = page.file_path.match(/content\/(news|statecollege)\//)
        ? "news-page"
        : page.file_path.match(/content\/newsletters/)
        ? "newsletter-page"
        : null;
      if (!name) {
        apiStateRefs.error.value = { name: "No admin associate with page." };
        return;
      }
      root.$router.push({
        name,
        params: {
          id: "" + page.id,
        },
      });
    });
    return {
      ...apiStateRefs,
      lookup,
      url,
    };
  },
});
</script>

<template>
  <form class="notification" @submit.prevent="lookup">
    <label class="label">Get page by URL</label>
    <div class="field is-grouped">
      <p class="control is-expanded">
        <input
          v-model="url"
          class="input"
          type="url"
          placeholder="https://www.spotlightpa.org/story/"
        />
      </p>
      <p class="control">
        <button
          class="button is-primary has-text-weight-semibold"
          :class="{ 'is-loading': isLoadingThrottled }"
          :disabled="isLoadingThrottled"
        >
          Search
        </button>
      </p>
      <p class="control">
        <BulmaPaste
          button-class="button is-success has-text-weight-semibold"
          @paste="
            url = $event;
            lookup();
          "
        />
      </p>
    </div>
    <p v-if="error" class="help is-danger" v-text="error.name"></p>
  </form>
</template>
