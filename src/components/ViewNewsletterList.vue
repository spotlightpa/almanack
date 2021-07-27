<script>
import { computed, toRefs, watch } from "@vue/composition-api";

import { formatDate } from "@/utils/time-format.js";

import { makeState } from "@/api/service-util.js";
import { useClient } from "@/api/client.js";

import APILoader from "./APILoader.vue";

class Page {
  constructor(data) {
    for (let [prop, key] of Object.entries({
      id: "id",
      internalID: "internal_id",
      title: "title",
      blurb: "blurb",
      description: "description",
      filePath: "file_path",
      urlPath: "url_path",
      createdAt: "created_at",
      lastPublished: "last_published",
      publishedAt: "published_at",
      updatedAt: "updated_at",
    })) {
      this[prop] = data[key];
    }
    for (let dateProp of ["createdAt", "publishedAt", "updatedAt"]) {
      if (this[dateProp]) {
        this[dateProp] = new Date(this[dateProp]);
      }
    }
  }
  get published() {
    return this.lastPublished.Valid;
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
    <h1 class="title">Newsletter Pages</h1>
    <APILoader :is-loading="isLoading" :reload="fetch" :error="error">
      <table class="table is-striped is-narrow is-fullwidth">
        <tbody>
          <tr v-for="page of pages" :key="page.id">
            <td>
              <a href class="is-block my-2">
                <span class="is-inline-flex middle">
                  <span class="tags mb-0">
                    <span
                      class="tag is-small has-text-weight-semibold"
                      :class="page.published ? 'is-success' : 'is-warning'"
                    >
                      <span class="icon is-size-6">
                        <font-awesome-icon
                          :icon="
                            page.published
                              ? ['fas', 'check-circle']
                              : ['fas', 'pen-nib']
                          "
                        />
                      </span>
                      <span>
                        {{ page.published ? "Published" : "Unpublished" }}
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
              </a>
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
