<script>
import { computed, watch } from "vue";
import { useClient, makeState } from "@/api/hooks.js";
import { Page } from "@/api/spotlightpa-page.js";

import useData from "@/utils/use-data.js";

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
      rawData.value ? new Page(rawData.value) : null
    );
    watch(
      () => props.item,
      (item) => load(item.page),
      { immediate: true }
    );

    return {
      ...apiStateRefs,
      page,

      ...useData(() => props.item, {
        label: ["label"],
        labelColor: ["labelColor"],
        bgColor: ["backgroundColor"],
        linkColor: ["linkColor"],
      }),
    };
  },
};
</script>

<template>
  <div>
    <h3 class="title is-4">
      #{{ pos + 1 }} {{ (page && page.internalID) || "" }}
    </h3>
    <details open>
      <summary>Settings</summary>

      <BulmaFieldInput v-model="label" label="Label for item" />
      <BulmaFieldColor v-model="bgColor" label="Item Background Color" />
      <BulmaFieldColor v-model="labelColor" label="Item Label Color" />
      <BulmaFieldColor
        v-model="linkColor"
        label="Item Link Color"
        help="Our orange is #ff6c36. Our yellow is #ffcb05. Our dark blue is #009edb. Our light blue is #99d9f1."
      />
    </details>
    <details open>
      <summary>Preview</summary>
      <div v-if="page">
        <div
          class="mt-4 plain-notification"
          data-ga-category="editors-pick"
          :style="{ backgroundColor: bgColor }"
        >
          <h1
            class="title is-size-5 is-size-4-mobile has-text-centered is-uppercase"
            :style="{ color: labelColor }"
          >
            {{ label }}
          </h1>

          <article class="mb-5">
            <figure>
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
                class="hover-underline"
                :style="{ color: linkColor }"
                :href="page.link"
                target="_blank"
              >
                {{ page.title }}
              </a>
            </h2>

            <h3
              v-if="page.byline"
              class="subtitle has-margin-top-negative-medium is-5 has-text-weight-normal"
            >
              <a
                :href="page.link"
                :style="{ color: linkColor }"
                class="hover-underline"
                target="_blank"
              >
                by
                {{ page.byline }}
              </a>
            </h3>
          </article>
        </div>
      </div>
    </details>
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
