<script setup>
import { computed } from "vue";

import imgproxyURL from "@/api/imgproxy-url.js";
import maybeDate from "@/utils/maybe-date.js";

const props = defineProps({
  modelValue: {
    type: Object,
    required: true,
  },
});

class TaxonomyPage {
  constructor(data) {
    this.id = data["id"] ?? "";
    this.filePath = data["file_path"] ?? "";
    this.frontmatter = data["frontmatter"] ?? {};
    this.body = data["body"] ?? "";
    this.createdAt = data["created_at"] ?? "";
    this.updatedAt = maybeDate(data, "updated_at");
    this.lastPublished = maybeDate(data, "last_published");
    this.publicationDate = maybeDate(this.frontmatter, "published");
    this.title = this.frontmatter["title"] ?? "";
    this.blurb = this.frontmatter["blurb"] ?? "";
    this.kicker = this.frontmatter["kicker"] ?? "";
    this.byline = this.frontmatter["byline"] ?? "";
    this.link = this.frontmatter["link"] ?? "";
    this.linktitle = this.frontmatter["linktitle"] ?? "";
    this.image = this.frontmatter["image"] ?? "";
    this.imageGravity = this.frontmatter["image-gravity"] ?? "";
    this.imageDescription = this.frontmatter["image-description"] ?? "";
    this.internalID = this.frontmatter["internal-id"] ?? "";
    this.draft = this.frontmatter["draft"] ?? false;
  }
  get thumbnailURL() {
    return imgproxyURL(this.image, {
      width: 256,
      height: 192,
      extension: "webp",
      gravity: this.imageGravity,
    });
  }
  get editLink() {
    return {
      name: "topic-page",
      params: {
        id: "" + this.id,
      },
    };
  }
}

const page = computed(() => new TaxonomyPage(props.modelValue));
</script>

<template>
  <router-link
    class="is-flex-tablet my-2 is-align-items-center"
    :to="page.editLink"
  >
    <div class="is-flex-grow-1">
      <p class="mt-0 has-text-weight-semibold has-text-black">
        {{ page.title }}
      </p>
      <p class="has-text-weight-light has-text-dark">
        {{ page.linktitle }}
      </p>
    </div>
    <div
      v-if="page.image"
      class="m-2 is-flex-shrink-0 is-clipped"
      style="width: 128px"
    >
      <picture class="image has-ratio">
        <img
          class="is-3x4"
          :src="page.thumbnailURL"
          :alt="page.imageDescription"
          loading="lazy"
        />
      </picture>
    </div>
  </router-link>
</template>
