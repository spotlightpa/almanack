<script setup>
import { computed, ref } from "vue";

import { post, postPageJSON } from "@/api/client-v2.js";
import { makeState } from "@/api/service-util.js";
import imgproxyURL from "@/api/imgproxy-url.js";
import maybeDate from "@/utils/maybe-date.js";

const props = defineProps({
  modelValue: {
    type: Object,
    required: true,
  },
});

class VideoPage {
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
    this.youtubeID = this.frontmatter["youtube-id"] ?? "";
    this.videoURL = this.frontmatter["video-url"] ?? "";
    this.videoType = this.frontmatter["video-type"] ?? "";
    this.image = this.frontmatter["image"] ?? "";
    this.imageDescription = this.frontmatter["image-description"] ?? "";
    this.internalID = this.frontmatter["internal-id"] ?? "";
    this.draft = this.frontmatter["draft"] ?? false;
  }

  get isPublished() {
    return !!this.lastPublished;
  }

  get status() {
    return this.isPublished ? "pub" : "none";
  }

  get isShort() {
    return this.videoType === "youtube-short";
  }

  get thumbnailURL() {
    return imgproxyURL(this.image, {
      width: 256,
      height: 192,
      extension: "webp",
    });
  }
}

const emit = defineEmits(["update:modelValue"]);

const video = computed(() => new VideoPage(props.modelValue));

const isOpen = ref(false);
let hasOpened = false;

const draft = ref(false);
const internalID = ref("");
const title = ref("");
const blurb = ref("");
const kicker = ref("");
const image = ref("");
const imageDescription = ref("");
const link = ref("");

function initValues() {
  draft.value = video.value.draft;
  internalID.value = video.value.internalID;
  title.value = video.value.title;
  blurb.value = video.value.blurb;
  kicker.value = video.value.kicker;
  image.value = video.value.image;
  imageDescription.value = video.value.imageDescription;
  link.value = video.value.link;
}

function toggle() {
  isOpen.value = !isOpen.value;
  if (!hasOpened) {
    hasOpened = true;
    initValues();
  }
}

const { exec, apiStateRefs } = makeState();

const { isLoadingThrottled, error } = apiStateRefs;

async function saveVideo() {
  let obj = {
    id: video.value.id,
    set_frontmatter: true,
    frontmatter: {
      ...video.value.frontmatter,
      draft: draft.value,
      "internal-id": internalID.value,
      title: title.value,
      blurb: blurb.value,
      kicker: kicker.value,
      image: image.value,
      "image-description": imageDescription.value,
      link: link.value,
    },
    set_body: false,
    set_schedule_for: false,
    schedule_for: null,
    set_last_published: true,
  };
  await exec(() => post(postPageJSON, obj));
  if (!apiStateRefs.error.value) {
    emit("update:modelValue", apiStateRefs.rawData.value);
  }
}
</script>

<template>
  <div class="is-flex-tablet my-2 is-align-items-center">
    <div class="is-flex-grow-1">
      <span class="is-inline-flex middle">
        <span class="tags mb-0">
          <TagDate :date="video.publicationDate"></TagDate>
          <a
            class="tag is-primary has-text-weight-semibold"
            target="_blank"
            :href="video.videoURL"
          >
            {{ video.videoType === "youtube-short" ? "Short" : "Video" }}
            {{ video.youtubeID }}
          </a>
        </span>
      </span>
      <p class="mt-0 has-text-weight-semibold has-text-black">
        {{ video.internalID }}
      </p>
      <p class="has-text-weight-light has-text-dark">
        {{ video.title }}
      </p>
      <p>
        <button
          class="mt-2 button is-light has-text-weight-semibold"
          type="button"
          v-if="!isOpen"
          @click="toggle"
        >
          Edit
        </button>
      </p>
    </div>
    <div
      v-if="video.image"
      class="m-2 is-flex-shrink-0 is-clipped"
      style="width: 128px"
    >
      <picture class="image has-ratio">
        <img
          class="is-3x4"
          :src="video.thumbnailURL"
          :alt="video.imageDescription"
          loading="lazy"
        />
      </picture>
    </div>
  </div>

  <KeepAlive>
    <div class="mb-3" v-if="isOpen">
      <BulmaFieldInput
        v-model="internalID"
        label="Slug"
        help="Slug used in Almanack as an internal ID"
      ></BulmaFieldInput>
      <BulmaFieldInput
        v-model="title"
        label="Hed"
        help="Used as link title"
      ></BulmaFieldInput>
      <BulmaFieldInput
        v-model="blurb"
        label="Blurb"
        help="Short summary to that appears as dek under the hed"
      ></BulmaFieldInput>
      <BulmaFieldInput
        v-model="kicker"
        label="Eyebrow"
        help='Small text appearing above the image thumbnail. Defaults to "Video" if blank.'
      ></BulmaFieldInput>
      <BulmaField
        label="Photo ID"
        help="Thumbnail shown in article curation"
        v-slot="{ idForLabel }"
      >
        <div class="is-flex">
          <input :id="idForLabel" v-model="image" class="input" />
          <BulmaPaste @paste="image = $event"></BulmaPaste>
        </div>
      </BulmaField>
      <BulmaTextarea
        v-model="imageDescription"
        label="SEO Image Alt Text"
      ></BulmaTextarea>

      <BulmaFieldInput
        v-model="link"
        label="Video link"
        type="url"
      ></BulmaFieldInput>

      <BulmaFieldCheckbox v-model="draft" label="Hide video">
        Remove video from lists on Spotlight PA
      </BulmaFieldCheckbox>

      <ErrorSimple :error="error"></ErrorSimple>
      <div class="buttons">
        <button
          class="button has-text-weight-semibold is-success"
          :class="{ 'is-loading': isLoadingThrottled }"
          type="button"
          @click="saveVideo"
        >
          <span> Save </span>
        </button>
        <button
          class="button has-text-weight-semibold is-danger"
          :disabled="isLoadingThrottled"
          type="button"
          @click="initValues"
        >
          <span> Discard Changes </span>
        </button>
        <button
          class="button has-text-weight-semibold is-light"
          type="button"
          @click="isOpen = false"
        >
          Close
        </button>
      </div>
    </div>
  </KeepAlive>
</template>
