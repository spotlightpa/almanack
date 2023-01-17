<script>
import { computed, watch } from "vue";

import { useClient, makeState } from "@/api/hooks.js";
import imgproxyURL from "@/api/imgproxy-url.js";

import { formatDate } from "@/utils/time-format.js";

const toImageObj = (rawImage) => ({
  id: rawImage.id,
  path: rawImage.path,
  url: imgproxyURL(rawImage.path, {
    width: 256,
    height: (256 * 9) / 16,
    extension: "webp",
  }),
  description: rawImage.description,
  credit: rawImage.credit,
  srcURL: rawImage.src_url,
  date: new Date(rawImage.created_at),
  downloadURL: "/ssr/download-image?src=" + encodeURIComponent(rawImage.path),
});

export default {
  props: { page: { type: String, default: "0" } },
  setup(props) {
    let { listImages, updateImage } = useClient();
    const { apiStateRefs, exec } = makeState();
    const { rawData } = apiStateRefs;

    const images = computed(() => {
      if (!rawData.value?.images) {
        return [];
      }
      return rawData.value.images.map((obj) => toImageObj(obj));
    });
    const nextPage = computed(() => {
      if (!rawData.value?.next_page) {
        return null;
      }
      return {
        name: "image-uploader",
        query: {
          page: "" + rawData.value.next_page,
        },
      };
    });

    let actions = {
      async fetch() {
        // Trigger image upload queue if not triggered already
        window.fetch("/api-background/images").catch(() => {});
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
      images,
      nextPage,
      ...actions,
      formatDate,
    };
  },
};
</script>

<template>
  <MetaHead>
    <title>Photo Uploads • Spotlight PA</title>
  </MetaHead>
  <div>
    <h1 class="title">
      Upload an image
      <template v-if="page !== '0'">(overflow page {{ page }})</template>
    </h1>

    <ImageUploader @update-image-list="fetch" />

    <h2 class="title has-margin-top">Existing Images</h2>
    <APILoader :is-loading="isLoading" :reload="fetch" :error="error">
      <table class="table is-striped is-fullwidth">
        <tbody>
          <tr v-for="image of images" :key="image.id">
            <td>
              <div class="max-128">
                <a :href="image.downloadURL" class="is-flex">
                  <picture
                    class="image is-3x4 has-background-grey-lighter has-margin-bottom"
                  >
                    <img
                      :src="image.url"
                      class="border-thick"
                      width="256"
                      height="192"
                      loading="lazy"
                    />
                  </picture>
                </a>
              </div>
              <div class="mt-4">
                <LinkHref
                  label="Original"
                  color="is-success"
                  :icon="['fas', 'file-download']"
                  :href="image.downloadURL"
                />
                <ImageSize class="mt-1" :path="image.path" />
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
  height: 72px;
  width: 128px;
  min-height: 1rem;
  min-width: 1rem;
}
.border-thick {
  border: 2px solid #ccc;
}
</style>
