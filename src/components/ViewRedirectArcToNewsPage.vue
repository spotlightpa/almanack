<script>
import { ref, toRefs, watch } from "@vue/composition-api";
import { makeState } from "@/api/service-util.js";
import { useClient } from "@/api/client.js";

export default {
  name: "ViewRedirectArcToNewsPage",
  props: {
    id: String,
  },
  metaInfo() {
    return {
      title: "Loading…",
    };
  },
  setup(props, { root }) {
    const needsCreation = ref(false);

    const isLoading = ref(false);
    window.setTimeout(() => {
      isLoading.value = true;
    }, 500);

    const { getPageForArcID, postPageForArcID } = useClient();
    const { apiState, exec } = makeState();

    const fetch = (id) =>
      exec(() => getPageForArcID({ params: { arc_id: id } }));

    watch(() => props.id, fetch, {
      immediate: true,
    });

    watch(
      () => apiState.rawData,
      (id) => {
        if (!id) {
          needsCreation.value = true;
          return;
        }
        root.$router.replace({
          name: "news-page",
          params: {
            id: "" + id,
          },
        });
      }
    );

    const { error } = toRefs(apiState);

    return {
      isDev: window.location.href.match(/localhost/),
      needsCreation,
      apiState,
      fetch,
      error,
      isLoading,

      convert(pagekind) {
        exec(() => postPageForArcID({ arc_id: props.id, page_kind: pagekind }));
      },
    };
  },
};
</script>

<template>
  <div>
    <div v-if="needsCreation">
      <h1 class="title">Convert article?</h1>
      <p>
        The article has not been converted from an Arc article to a Spotlight PA
        page. Convert now?
      </p>
      <p class="mt-2">
        <i>
          Note that after conversion, metadata such as article titles, images,
          and URL slugs must be manually updated to match Arc.
        </i>
      </p>
      <div class="mt-4 buttons">
        <button
          class="button is-primary has-text-weight-semibold"
          type="button"
          @click="convert('news')"
        >
          Convert as a News article
        </button>
        <button
          class="button is-primary has-text-weight-semibold"
          type="button"
          :disabled="!isDev"
          @click="convert('statecollege')"
        >
          Convert as a State College article
        </button>
      </div>
    </div>
    <SpinnerProgress :is-loading="isLoading && apiState.isLoading" />
    <ErrorReloader :error="error" @reload="fetch(id)" />
  </div>
</template>
