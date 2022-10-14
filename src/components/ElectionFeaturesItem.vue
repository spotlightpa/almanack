<script>
import { computed, reactive, watch } from "vue";
import { useClient, makeState } from "@/api/hooks.js";
import { Page } from "@/api/spotlightpa-page.js";

export default {
  props: { item: Object, pos: Number, length: Number },
  setup(props) {
    let { getPageByFilePath } = useClient();
    let { apiStateRefs, exec } = makeState();
    const { rawData } = apiStateRefs;

    async function load(path) {
      let params = { path, select: "-body" };
      return await exec(() => getPageByFilePath({ params }));
    }
    const page = computed(() =>
      rawData.value ? reactive(new Page(rawData.value)) : null
    );
    watch(
      () => props.item,
      (item) => load(item.page),
      { immediate: true }
    );

    return {
      ...apiStateRefs,
      page,
    };
  },
};
</script>

<template>
  <div>
    <h3 class="title is-4">
      #{{ pos + 1 }} {{ (page && page.internalID) || "" }}
    </h3>
    <article v-if="page" class="mb-5 plain-notification">
      <figure v-if="pos === 0">
        <picture v-if="page.imagePreviewURL" class="has-ratio">
          <img
            :src="page.getImagePreviewURL({ width: 480, height: 270 })"
            class="is-16x9"
            width="315"
          />
        </picture>
        <figcaption
          class="is-clearfix is-size-7 is-uppercase has-text-grey-light"
        >
          <span
            class="is-single-spaced is-pulled-right"
            v-text="page.imageCredit"
          >
          </span>
        </figcaption>
      </figure>
      <h2 class="title is-4 is-spaced mt-2">
        <a
          class="hover-underline has-text-black"
          :href="page.link"
          target="_blank"
        >
          {{ page.title }}
        </a>
      </h2>

      <h3
        v-if="page.byline && pos === 0"
        class="subtitle has-margin-top-negative-medium is-5 has-text-weight-normal"
      >
        <a
          :href="page.link"
          class="hover-underline has-text-black"
          target="_blank"
        >
          by
          {{ page.byline }}
        </a>
      </h3>
    </article>
    <div class="mt-2 buttons are-small">
      <button
        v-if="length > 1 && pos > 0"
        class="button has-text-weight-semibold is-success"
        type="button"
        @click="$emit('swap', { pos, dir: -1 })"
      >
        <span class="icon">
          <font-awesome-icon :icon="['fas', 'arrow-up']" />
        </span>
        <span>Move up</span>
      </button>
      <button
        v-if="length > 1 && pos < length - 1"
        class="button has-text-weight-semibold is-success"
        type="button"
        @click="$emit('swap', { pos, dir: 1 })"
      >
        <span class="icon">
          <font-awesome-icon :icon="['fas', 'arrow-down']" />
        </span>
        <span>Move down</span>
      </button>
      <button
        class="button has-text-weight-semibold is-danger"
        type="button"
        @click="$emit('remove', pos)"
      >
        <span class="icon">
          <font-awesome-icon :icon="['fas', 'trash-alt']" />
        </span>
        <span>Remove</span>
      </button>
    </div>
  </div>
</template>

<style scoped>
.plain-notification {
  max-width: 350px;
  border-radius: 4px;
  padding: 1.25rem 1.5rem 1.25rem 1.5rem;
}
</style>
