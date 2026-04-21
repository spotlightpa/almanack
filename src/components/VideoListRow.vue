<script setup>
import { ref } from "vue";
import { post, postPageJSON } from "@/api/client-v2.js";
import { makeState } from "@/api/service-util.js";

const props = defineProps({
  video: {
    type: Object,
    required: true,
  },
});
const emit = defineEmits(["refresh-list"]);

const isOpen = ref(false);
let hasOpened = false;

const internalID = ref("");
const title = ref("");
const blurb = ref("");
const kicker = ref("");
const image = ref("");
const imageDescription = ref("");
const link = ref("");

function initValues() {
  internalID.value = props.video.internalID;
  title.value = props.video.title;
  blurb.value = props.video.blurb;
  kicker.value = props.video.kicker;
  image.value = props.video.image;
  imageDescription.value = props.video.imageDescription;
  link.value = props.video.link;
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
    id: props.video.id,
    set_frontmatter: true,
    frontmatter: {
      ...props.video.frontmatter,
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
  emit("refresh-list", null);
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
            Video {{ video.youtubeID }}
          </a>
        </span>
      </span>
      <p class="mt-0 has-text-weight-semibold has-text-black">
        {{ video.title }}
      </p>
      <p class="has-text-weight-light has-text-dark">
        {{ video.description }}
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
