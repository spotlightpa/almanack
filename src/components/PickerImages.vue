<script>
import imgproxyURL from "@/api/imgproxy-url.js";

export default {
  props: { images: Array },
  setup() {
    return {
      imgproxyURL,
    };
  },
};
</script>

<template>
  <BulmaField v-if="images.length" label="Choose from recent photos">
    <div class="textarea preview-frame">
      <table class="table is-striped is-narrow is-fullwidth">
        <tbody>
          <tr v-for="image in images" :key="image.id">
            <a
              class="is-flex-tablet p-1 has-text-black"
              @click="$emit('select-image', image)"
            >
              <div
                class="mr-2 is-flex-shrink-0 is-clipped"
                style="width: 128px"
              >
                <picture class="has-ratio">
                  <img
                    class="is-3x4"
                    :src="
                      imgproxyURL(image.path, {
                        width: 256,
                        height: 192,
                        extension: 'webp',
                      })
                    "
                    :alt="image.path"
                    loading="lazy"
                  />
                </picture>
              </div>
              <div>
                <div class="clamped-3">
                  {{ image.description }}
                  <template v-if="image.credit">
                    ({{ image.credit }})
                  </template>
                </div>
              </div>
            </a>
          </tr>
        </tbody>
      </table>
    </div>
    <p>
      <router-link :to="{ name: 'uploader' }" target="_blank">
        Manage photos
      </router-link>
    </p>
  </BulmaField>
</template>

<style scoped>
.clamped-3 {
  display: -webkit-box;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 3;
  overflow: hidden;
}
.preview-frame {
  height: 300px;
  overflow-y: auto;
}
</style>
