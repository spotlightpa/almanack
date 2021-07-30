<script>
import { computed, toRefs, watch } from "@vue/composition-api";

import { formatDate } from "@/utils/time-format.js";

import { makeState } from "@/api/service-util.js";
import { useClient } from "@/api/client.js";
import imgproxyURL from "@/api/imgproxy-url.js";

import APILoader from "./APILoader.vue";

class Page {
  constructor(data) {
    this.id = data["id"] || "";
    this.internalID = data["internal_id"] || "";
    this.title = data["title"] || "";
    this.blurb = data["blurb"] || "";
    this.description = data["description"] || "";
    this.filePath = data["file_path"] || "";
    this.urlPath = data["url_path"] || "";
    this.image = data["image"] || "";
    this.createdAt = Page.getDate(data, "created_at");
    this.publishedAt = Page.getDate(data, "published_at");
    this.updatedAt = Page.getDate(data, "updated_at");
    this.lastPublished = Page.getNullableDate(data, "last_published");
    this.scheduleFor = Page.getNullableDate(data, "schedule_for");
  }

  static getDate(data, prop) {
    let date = data[prop] ?? null;
    return date && new Date(date);
  }

  static getNullableDate(data, prop) {
    return data[prop]?.Valid ? new Date(data[prop].Time) : null;
  }

  get isPublished() {
    return !!this.lastPublished;
  }

  get status() {
    if (this.isPublished) {
      return "pub";
    }
    return this.scheduleFor ? "sked" : "none";
  }

  get statusVerbose() {
    return {
      pub: "Published",
      sked: "Scheduled",
      none: "Unpublished",
    }[this.status];
  }

  get statusClass() {
    return {
      pub: "is-success",
      sked: "is-warning",
      none: "is-danger",
    }[this.status];
  }

  get link() {
    return {
      name: "newsletter-page",
      params: {
        id: "" + this.id,
      },
    };
  }

  get imgproxyURL() {
    return imgproxyURL(this.image, {
      width: 256,
      height: 192,
      extension: "webp",
    });
  }
}

export default {
  name: "ViewNewsletterList",
  components: { APILoader },
  props: ["page"],
  metaInfo: {
    title: "Newsletter Pages",
  },
  setup(props) {
    let { listNewsletterPages } = useClient();
    let { apiState, exec } = makeState();

    const fetch = (page) => exec(() => listNewsletterPages(page));

    watch(() => props.page, fetch, {
      immediate: true,
    });

    return {
      formatDate,

      ...toRefs(apiState),
      fetch,
      pages: computed(() => {
        let pages = apiState.rawData?.pages;
        if (!pages) return [];
        return pages.map((page) => new Page(page));
      }),
      nextPage: computed(() => {
        let param = apiState.rawData?.next_page;
        if (!param) return null;
        return {
          name: "newsletters",
          query: {
            page: param,
          },
        };
      }),
    };
  },
};
</script>

<template>
  <div>
    <nav class="breadcrumb has-succeeds-separator" aria-label="breadcrumbs">
      <ul>
        <li>
          <router-link :to="{ name: 'admin' }">Admin</router-link>
        </li>
        <li class="is-active">
          <router-link exact :to="{ name: 'newsletters' }">
            Newsletter Pages
          </router-link>
        </li>
      </ul>
    </nav>

    <h1 class="title">
      Newsletter Pages
      <template v-if="page">(overflow page {{ page }})</template>
    </h1>
    <APILoader :is-loading="isLoading" :reload="fetch" :error="error">
      <table class="table is-striped is-narrow is-fullwidth">
        <tbody>
          <tr v-for="page of pages" :key="page.id">
            <td>
              <router-link class="is-flex-tablet my-2" :to="page.link">
                <div class="is-flex-grow-1">
                  <span class="is-inline-flex middle">
                    <span class="tags mb-0">
                      <span
                        class="tag is-small has-text-weight-semibold"
                        :class="page.statusClass"
                      >
                        <span class="icon is-size-6">
                          <font-awesome-icon
                            :icon="
                              page.isPublished
                                ? ['fas', 'check-circle']
                                : ['fas', 'pen-nib']
                            "
                          />
                        </span>
                        <span>
                          {{ page.statusVerbose }}
                        </span>
                      </span>
                      <span
                        class="tag is-primary has-text-weight-semibold"
                        v-text="page.internalID"
                      ></span>
                      <span class="tag is-light has-text-weight-semibold">
                        {{ formatDate(page.publishedAt) }}
                      </span>
                    </span>
                  </span>
                  <p class="mt-0 has-text-weight-bold has-text-black">
                    {{ page.title }}
                  </p>
                  <p class="has-text-weight-light has-text-dark">
                    {{ page.blurb }}
                  </p>
                </div>
                <div
                  v-if="page.image"
                  class="is-flex-grow-0 is-clipped"
                  style="width: 128px"
                >
                  <picture class="has-ratio">
                    <img
                      class="is-3x4"
                      :src="page.imgproxyURL"
                      :alt="page.image"
                      loading="lazy"
                    />
                  </picture>
                </div>
              </router-link>
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
          Show Older Stories…
        </router-link>
      </div>
    </APILoader>
  </div>
</template>
