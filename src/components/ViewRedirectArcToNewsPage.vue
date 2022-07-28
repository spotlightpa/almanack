<script>
import { inject, ref, watch } from "vue";
import { makeState } from "@/api/service-util.js";
import { useClient } from "@/api/client.js";

export default {
  name: "ViewRedirectArcToNewsPage",
  props: {
    id: String,
  },
  setup(props) {
    const router = inject("router");
    const needsCreation = ref(false);

    const isLoadingDebounced = ref(false);
    window.setTimeout(() => {
      isLoadingDebounced.value = true;
    }, 500);

    const { getPageForArcID, postPageForArcID } = useClient();
    const { apiStateRefs, exec } = makeState();

    const fetch = (id) =>
      exec(() => getPageForArcID({ params: { arc_id: id } }));

    watch(() => props.id, fetch, {
      immediate: true,
    });

    watch(apiStateRefs.rawData, (id) => {
      if (!id) {
        needsCreation.value = true;
        return;
      }
      router.replace({
        name: "news-page",
        params: {
          id: "" + id,
        },
      });
    });

    return {
      ...apiStateRefs,
      isLoadingDebounced,
      isDev: window.location.href.match(/localhost/),
      needsCreation,
      fetch,

      convert(pagekind) {
        exec(() => postPageForArcID({ arc_id: props.id, page_kind: pagekind }));
      },
    };
  },
};
</script>

<template>
  <div>
    <MetaHead>
      <title>Loading…</title>
    </MetaHead>
    <div v-if="needsCreation">
      <MetaHead>
        <title>Convert Page • Spotlight PA Almanack</title>
      </MetaHead>

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
          @click="convert('statecollege')"
        >
          Convert as a State College article
        </button>
      </div>
    </div>
    <SpinnerProgress :is-loading="isLoadingDebounced && isLoading" />
    <ErrorReloader :error="error" @reload="fetch(id)" />
  </div>
</template>
