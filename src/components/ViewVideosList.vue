<script setup>
// import { reactive, ref } from "vue";

import { watchAPI } from "@/api/service-util.js";
import { get, listPages } from "@/api/client-v2.js";
import imgproxyURL from "@/api/imgproxy-url.js";
import maybeDate from "@/utils/maybe-date.js";
import VideoListRow from "./VideoListRow.vue";

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
    this.link = this.frontmatter["link"] ?? "";
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

const videos = computedList("pages", (page) => new VideoPage(page));
const nextPage = computedProp("next_page", (page) => ({
  name: "video-pages",
  query: { page },
}));
</script>

<template>
  <MetaHead>
    <title>Videos • Spotlight PA Almanack</title>
  </MetaHead>

  <div>
    <BulmaBreadcrumbs
      :links="[
        { name: 'Admin', to: { name: 'admin' } },
        { name: 'Videos', to: {} },
      ]"
    ></BulmaBreadcrumbs>
    <h1 class="title">
      Videos
      <template v-if="page">(overflow page {{ page }})</template>
    </h1>

    <APILoader
      :is-loading="apiState.isLoading.value"
      :reload="fetch"
      :error="apiState.error.value"
    >
      <table class="table is-striped is-narrow is-fullwidth">
        <tbody>
          <tr v-for="video of videos" :key="video.id">
            <td>
              <VideoListRow :video="video" @refresh-list="fetch" />
            </td>
          </tr>
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

<style scoped>
.zebra-row {
  background-color: #fff;
}

.zebra-row:nth-child(even) {
  background-color: #fafafa;
}

.zebra-row + .zebra-row {
  border-top: 1px solid #dbdbdb;
}
</style>
