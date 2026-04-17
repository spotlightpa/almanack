<script setup>
import { reactive, ref } from "vue";

import { watchAPI } from "@/api/service-util.js";
import { get, post, listPages, postPageJSON } from "@/api/client-v2.js";
import imgproxyURL from "@/api/imgproxy-url.js";
import maybeDate from "@/utils/maybe-date.js";

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
    this.description = this.frontmatter["description"] ?? "";
    this.blurb = this.frontmatter["blurb"] ?? "";
    this.kicker = this.frontmatter["kicker"] ?? "";
    this.byline = this.frontmatter["byline"] ?? "";
    this.youtubeID = this.frontmatter["youtube-id"] ?? "";
    this.videoURL = this.frontmatter["video-url"] ?? "";
    this.videoType = this.frontmatter["video-type"] ?? "";
    this.image = this.frontmatter["image"] ?? "";
    this.imageDescription = this.frontmatter["image-description"] ?? "";
    this.internalID = this.frontmatter["internal-id"] ?? "";
  }

  get isPublished() {
    return !!this.lastPublished;
  }

  get status() {
    return this.isPublished ? "pub" : "none";
  }

  get youtubeURL() {
    if (this.videoURL) return this.videoURL;
    if (this.youtubeID)
      return `https://www.youtube.com/watch?v=${this.youtubeID}`;
    return "";
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

  toJSON() {
    return {
      id: this.id,
      set_frontmatter: true,
      frontmatter: {
        ...this.frontmatter,
        title: this.title,
        description: this.description,
        blurb: this.blurb,
        kicker: this.kicker,
        byline: this.byline,
        "image-description": this.imageDescription,
      },
      set_body: true,
      body: this.body,
      set_schedule_for: false,
      schedule_for: null,
      url_path: "",
      set_last_published: false,
    };
  }
}

const props = defineProps({
  page: { default: "" },
});

const { apiState, fetch, computedList, computedProp } = watchAPI(
  () => props.page,
  (page) =>
    get(listPages, {
      page,
      path: "content/videos/",
      select: "frontmatter",
    })
);

const videos = computedList("pages", (page) => reactive(new VideoPage(page)));
const nextPage = computedProp("next_page", (page) => ({
  name: "video-pages",
  query: { page },
}));

const expandedRow = ref(null);
function toggleRow(id) {
  expandedRow.value = expandedRow.value === id ? null : id;
}

const savingIds = reactive(new Set());

async function saveVideo(video) {
  savingIds.add(video.id);
  let [data, err] = await post(postPageJSON, video);
  savingIds.delete(video.id);
  if (err) {
    window.alert("Error saving: " + err.message);
    return;
  }
  // Update the video with new data from the server
  Object.assign(video, new VideoPage(data));
}
</script>

<template>
  <MetaHead>
    <title>YouTube Videos • Spotlight PA Almanack</title>
  </MetaHead>

  <div>
    <BulmaBreadcrumbs
      :links="[
        { name: 'Admin', to: { name: 'admin' } },
        { name: 'YouTube Videos', to: {} },
      ]"
    ></BulmaBreadcrumbs>
    <h1 class="title">
      YouTube Videos
      <template v-if="page">(overflow page {{ page }})</template>
    </h1>

    <APILoader
      :is-loading="apiState.isLoading.value"
      :reload="fetch"
      :error="apiState.error.value"
    >
      <table class="table is-striped is-fullwidth">
        <thead>
          <tr>
            <th style="width: 140px">Thumbnail</th>
            <th>Title</th>
            <th style="width: 100px">Type</th>
            <th style="width: 100px">Status</th>
            <th style="width: 120px">Date</th>
            <th style="width: 80px"></th>
          </tr>
        </thead>
        <tbody>
          <template v-for="video of videos" :key="video.id">
            <tr class="is-clickable" @click="toggleRow(video.id)">
              <td>
                <picture v-if="video.image" class="image" style="width: 128px">
                  <img
                    :src="video.thumbnailURL"
                    :alt="video.imageDescription"
                    loading="lazy"
                    class="is-3x4"
                  />
                </picture>
              </td>
              <td>
                <p class="has-text-weight-semibold">{{ video.title }}</p>
                <p class="is-size-7 has-text-grey">
                  <a
                    v-if="video.youtubeURL"
                    :href="video.youtubeURL"
                    target="_blank"
                    @click.stop
                  >
                    <span class="icon is-small">
                      <font-awesome-icon
                        :icon="['fas', 'external-link-alt']"
                      ></font-awesome-icon>
                    </span>
                    YouTube
                  </a>
                </p>
              </td>
              <td>
                <span
                  class="tag is-small"
                  :class="video.isShort ? 'is-info' : 'is-dark'"
                >
                  {{ video.isShort ? "Short" : "Video" }}
                </span>
              </td>
              <td>
                <TagStatus :status="video.status"></TagStatus>
              </td>
              <td>
                <TagDate :date="video.publicationDate"></TagDate>
              </td>
              <td>
                <button
                  class="button is-small is-ghost"
                  @click.stop="toggleRow(video.id)"
                >
                  <span class="icon">
                    <font-awesome-icon
                      :icon="
                        expandedRow === video.id
                          ? ['fas', 'chevron-up']
                          : ['fas', 'chevron-down']
                      "
                    ></font-awesome-icon>
                  </span>
                </button>
              </td>
            </tr>

            <!-- Inline editor -->
            <tr v-if="expandedRow === video.id">
              <td colspan="6">
                <div class="box">
                  <div class="columns">
                    <div class="column">
                      <BulmaFieldInput
                        v-model="video.title"
                        label="Title"
                      ></BulmaFieldInput>
                      <BulmaFieldInput
                        v-model="video.kicker"
                        label="Kicker"
                      ></BulmaFieldInput>
                      <BulmaFieldInput
                        v-model="video.byline"
                        label="Byline"
                      ></BulmaFieldInput>
                    </div>
                    <div class="column">
                      <BulmaField label="Description">
                        <BulmaTextarea
                          v-model="video.description"
                          :rows="3"
                        ></BulmaTextarea>
                      </BulmaField>
                      <BulmaField label="Blurb">
                        <BulmaTextarea
                          v-model="video.blurb"
                          :rows="2"
                        ></BulmaTextarea>
                      </BulmaField>
                      <BulmaFieldInput
                        v-model="video.imageDescription"
                        label="Image Description"
                      ></BulmaFieldInput>
                    </div>
                  </div>
                  <div class="field is-grouped">
                    <div class="control">
                      <button
                        class="button is-success has-text-weight-semibold"
                        :class="{ 'is-loading': savingIds.has(video.id) }"
                        @click.stop="saveVideo(video)"
                      >
                        <span class="icon">
                          <font-awesome-icon
                            :icon="['fas', 'save']"
                          ></font-awesome-icon>
                        </span>
                        <span>Save &amp; Publish</span>
                      </button>
                    </div>
                    <div class="control">
                      <button
                        class="button is-light"
                        @click.stop="toggleRow(null)"
                      >
                        Cancel
                      </button>
                    </div>
                  </div>
                </div>
              </td>
            </tr>
          </template>
        </tbody>
      </table>

      <div class="buttons mt-5">
        <router-link
          v-if="nextPage"
          :to="nextPage"
          class="button is-primary has-text-weight-semibold"
        >
          Show More Videos…
        </router-link>
      </div>
    </APILoader>
  </div>
</template>
