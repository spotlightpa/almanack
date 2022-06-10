<script>
import { reactive, computed, toRefs, watch } from "@vue/composition-api";

import { useClient, makeState } from "@/api/hooks.js";
import imgproxyURL from "@/api/imgproxy-url.js";

import fuzzyMatch from "@/utils/fuzzy-match.js";
import { formatDate } from "@/utils/time-format.js";

const toImageObj = (rawImage) => ({
  id: rawImage.id,
  path: rawImage.path,
  url: imgproxyURL(rawImage.path, {
    width: 256,
    height: 192,
    extension: "webp",
  }),
  description: rawImage.description,
  credit: rawImage.credit,
  srcURL: rawImage.src_url,
  date: new Date(rawImage.created_at),
});
const imageProps = (image) => [image.description, image.credit];

export default {
  props: { page: { type: String, default: "0" } },
  metaInfo: {
    title: "Photos",
  },
  setup(props) {
    let { listImages, updateImage } = useClient();
    const { apiStateRefs, exec } = makeState();
    const { rawData } = apiStateRefs;

    let state = reactive({
      images: computed(() => {
        if (!rawData.value?.images) {
          return [];
        }
        return rawData.value.images.map((obj) => toImageObj(obj));
      }),
      rawFilter: "",
      imageProps: computed(() =>
        Array.from(
          new Set(state.images.flatMap((article) => imageProps(article)))
        ).sort()
      ),
      filteredImages: computed(() => {
        if (!state.rawFilter) {
          return state.images;
        }
        return state.images.filter((article) =>
          imageProps(article).some((prop) => fuzzyMatch(prop, state.rawFilter))
        );
      }),
      filterOptions: computed(() =>
        state.imageProps.filter((prop) => fuzzyMatch(prop, state.rawFilter))
      ),
      nextPage: computed(() => {
        if (!rawData.value?.next_page) {
          return null;
        }
        return {
          name: "image-uploader",
          query: {
            page: "" + rawData.value.next_page,
          },
        };
      }),
    });

    let actions = {
      async fetch() {
        return exec(() => listImages({ params: { page: props.page } }));
      },
      updateDescription(image) {
        let description = window.prompt(
          "Update description",
          image.description
        );
        if (description !== null && description !== image.description) {
          actions.doUpdate(image, { description });
        }
      },
      updateCredit(image) {
        let credit = window.prompt("Update credit", image.credit);
        if (credit !== null && credit !== image.credit) {
          actions.doUpdate(image, { credit });
        }
      },
      async doUpdate(image, opt) {
        return exec(async () => {
          await updateImage(image.path, opt);
          return listImages({ params: { page: props.page } });
        });
      },
    };

    watch(
      () => props.page,
      () => actions.fetch(),
      { immediate: true }
    );

    return {
      ...apiStateRefs,
      ...toRefs(state),
      ...actions,
      formatDate,
    };
  },
};
</script>

<template>
  <div>
    <h1 class="title">
      Upload an image
      <template v-if="page !== '0'">(overflow page {{ page }})</template>
    </h1>

    <ImageUploader @update-image-list="fetch" />

    <h2 class="title has-margin-top">Existing Images</h2>
    <APILoader :is-loading="isLoading" :reload="fetch" :error="error">
      <b-field label="">
        <b-autocomplete
          v-model="rawFilter"
          :data="filterOptions"
          placeholder="Filter images"
          clearable
        >
          <template v-slot:empty>No results found</template>
        </b-autocomplete>
      </b-field>

      <table class="table is-striped is-fullwidth">
        <tbody>
          <tr v-for="image of filteredImages" :key="image.id">
            <td>
              <div class="max-128">
                <picture
                  class="image is-3x4 has-background-grey-lighter has-margin-bottom"
                >
                  <img
                    :src="image.url"
                    width="256"
                    height="192"
                    loading="lazy"
                  />
                </picture>
              </div>
            </td>
            <td>
              <p>
                <a
                  class="has-text-weight-semibold"
                  @click="updateDescription(image)"
                >
                  Description:
                </a>
                {{ image.description }}
              </p>

              <p>
                <a
                  class="has-text-weight-semibold"
                  @click="updateCredit(image)"
                >
                  Credit:
                </a>
                {{ image.credit }}
              </p>
              <p>
                <strong>Date:</strong>
                {{ formatDate(image.date) }}
              </p>
              <p class="has-margin-top-thin">
                <CopyWithButton
                  :value="image.path"
                  label="path"
                  size="is-small"
                />
              </p>
            </td>
            <td></td>
          </tr>
        </tbody>
      </table>
      <div class="buttons mt-5">
        <router-link
          v-if="nextPage"
          :to="nextPage"
          class="button is-primary has-text-weight-semibold"
        >
          Show Older Images…
        </router-link>
      </div>
    </APILoader>
  </div>
</template>

<style scoped>
.max-128 {
  height: 128px;
  width: 128px;
  min-height: 1rem;
  min-width: 1rem;
}
</style>
