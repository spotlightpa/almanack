<script>
import { reactive, computed, toRefs } from "@vue/composition-api";

import { useClient } from "@/api/hooks.js";
import imgproxyURL from "@/api/imgproxy-url.js";
import fuzzyMatch from "@/utils/fuzzy-match.js";

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
    title: "Images",
  },
  setup() {
    let { listImages } = useClient();

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
          url: imgproxyURL(rawImage.path, { width: 256, height: 256 }),
          description: rawImage.description,
          credit: rawImage.credit,
          srcURL: rawImage.src_url,
          date: new Date(rawImage.created_at),
        }));
      },
    };

    actions.fetch();

    return {
      ...toRefs(state),
      ...actions,
    };
  },
};
</script>

<template>
  <div>
    <h1 class="title">Upload an image</h1>

    <ImageUploader />

    <h2 class="title has-margin-top">Existing Images</h2>
    <APILoader
      :can-load="true"
      :is-loading="isLoading"
      :reload="fetch"
      :error="error"
    >
      <b-field label="">
        <b-autocomplete
          v-model="rawFilter"
          :data="filterOptions"
          placeholder="Filter images"
          clearable
        >
          <template slot="empty">No results found</template>
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
                <strong>Description:</strong>
                {{ image.description }}
              </p>

              <p>
                <strong>Credit:</strong>
                {{ image.credit }}
              </p>
              <p>
                <strong>Date:</strong>
                {{ image.date | formatDate }}
              </p>
              <p class="has-margin-top-thin">
                <CopyWithButton
                  :value="image.path"
                  label="path"
                  size="is-small"
                ></CopyWithButton>
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
