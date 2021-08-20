<script>
import imgproxyURL from "@/api/imgproxy-url.js";

import TagDate from "./TagDate.vue";
import TagStatus from "./TagStatus.vue";

export default {
  name: "PageListRow",
  components: { TagStatus, TagDate },
  props: ["link", "status", "label", "date", "hed", "dek", "image", "imageAlt"],

  computed: {
    imgproxyURL() {
      return imgproxyURL(this.image, {
        width: 256,
        height: 192,
        extension: "webp",
      });
    },
  },
};
</script>

<template>
  <tr>
    <td>
      <router-link class="is-flex-tablet my-2" :to="link">
        <div class="is-flex-grow-1">
          <span class="is-inline-flex middle">
            <span class="tags mb-0">
              <TagStatus :status="status" />
              <span
                class="tag is-primary has-text-weight-semibold"
                v-text="label"
              ></span>
              <TagDate :date="date" />
            </span>
          </span>
          <p class="mt-0 has-text-weight-bold has-text-black">
            {{ hed }}
          </p>
          <p class="has-text-weight-light has-text-dark">
            {{ dek }}
          </p>
        </div>
        <div
          v-if="image"
          class="is-flex-grow-0 is-clipped"
          style="width: 128px"
        >
          <picture class="has-ratio">
            <img
              class="is-3x4"
              :src="imgproxyURL"
              :alt="imageAlt"
              loading="lazy"
            />
          </picture>
        </div>
      </router-link>
    </td>
  </tr>
</template>
