<script>
import { reactive, computed, toRefs } from "@vue/composition-api";

import { useClient } from "@/api/hooks.js";
import imgproxyURL from "@/api/imgproxy-url.js";
import fuzzyMatch from "@/utils/fuzzy-match.js";
import { formatDate } from "@/utils/time-format.js";

import APILoader from "./APILoader.vue";
import CopyWithButton from "./CopyWithButton.vue";
import ImageUploader from "./ImageUploader.vue";

export default {
  name: "ViewUploader",
  components: {
    APILoader,
    CopyWithButton,
    ImageUploader,
  },
  metaInfo: {
    title: "Photos",
  },
  setup() {
    let { listImages, updateImage } = useClient();

    const imageProps = (image) => [image.description, image.credit];

    let state = reactive({
      isLoading: false,
      images: [],
      error: null,
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
    });

    let actions = {
      async fetch() {
        state.isLoading = true;
        let data;
        [data, state.error] = await listImages();
        state.isLoading = false;
        if (state.error) {
          return;
        }
        state.images = data.images.map((rawImage) => ({
          id: rawImage.id,
          path: rawImage.path,
          url: imgproxyURL(rawImage.path, {
            width: 256,
            height: 256,
            extension: rawImage.type,
          }),
          description: rawImage.description,
          credit: rawImage.credit,
          srcURL: rawImage.src_url,
          date: new Date(rawImage.created_at),
        }));
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
        state.isLoading = true;
        await updateImage(image.path, opt);
        await actions.fetch();
      },
    };

    actions.fetch();

    return {
      ...toRefs(state),
      ...actions,

      formatDate,
    };
  },
};
</script>

<template>
  <div>
    <h1 class="title">Upload an image</h1>

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
                  class="image is-square has-background-grey-lighter has-margin-bottom"
                >
                  <img :src="image.url" width="256" height="256" />
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
