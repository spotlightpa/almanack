<script>
import { reactive, toRefs } from "@vue/composition-api";

import { useClient } from "@/api/hooks.js";
import imgproxyURL from "@/api/imgproxy-url.js";

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

    let state = reactive({
      isLoading: false,
      images: [],
      error: null,
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

    <h2 class="title has-margin-top">Images</h2>
    <APILoader
      :can-load="true"
      :is-loading="isLoading"
      :reload="fetch"
      :error="error"
    >
      <figure v-for="image of images" :key="image.id">
        <div class="max-128">
          <picture
            class="image is-square has-background-grey-lighter has-margin-bottom"
          >
            <img :src="image.url" width="256" height="256" />
          </picture>
        </div>
        <figcaption>
          <p class="has-margin-bottom-thin">
            <strong>Description:</strong>
          </p>
          <CopyWithButton
            :value="image.description"
            label="description"
          ></CopyWithButton>
          <p class="has-margin-bottom-thin">
            <strong>Credit:</strong>
          </p>
          <CopyWithButton :value="image.credit" label="credit"></CopyWithButton>
        </figcaption>
      </figure>
    </APILoader>
  </div>
</template>

<style scoped>
.max-128 {
  max-height: 128px;
  max-width: 128px;
  min-height: 1rem;
  min-width: 1rem;
}
</style>
