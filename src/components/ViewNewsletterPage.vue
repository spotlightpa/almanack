<script>
import { computed, toRefs, watch } from "@vue/composition-api";

import { makeState } from "@/api/service-util.js";
import { useClient } from "@/api/client.js";

class Page {
  constructor(data) {
    this.init(data);
  }

  init(data) {
    this.id = data["id"] ?? "";
    this.body = data["body"] ?? "";
    this.frontmatter = data["frontmatter"] ?? "";
    this.filePath = data["file_path"] ?? "";
    this.urlPath = data["url_path"]?.String ?? "";
    this.createdAt = data["created_at"] ?? "";
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

  get internalID() {
    return this.frontmatter["internal-id"] || "";
  }
  get title() {
    return "Page " + (this.internalID || this.frontmatter.title || "Untitled");
  }
  get published() {
    return !!this.lastPublished;
  }
  get link() {
    return {
      name: "newsletter-page",
      params: {
        id: this.id,
      },
    };
  }
}

export default {
  name: "ViewNewsletterPage",
  components: {},
  props: {
    id: String,
  },
  metaInfo() {
    return {
      title: this.title,
    };
  },
  setup(props) {
    let { getPage } = useClient();
    let { apiState, exec } = makeState();

    const fetch = (id) => exec(() => getPage(id));

    watch(() => props.id, fetch, {
      immediate: true,
    });
    const page = computed(() =>
      apiState.rawData ? new Page(apiState.rawData) : null
    );

    return {
      ...toRefs(apiState),

      fetch,
      page,
      title: computed(() => {
        if (page.value) {
          return page.value.title;
        }
        return `Newsletter Page ${props.id}`;
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
        <li>
          <router-link exact :to="{ name: 'newsletters' }">
            Newsletter Pages
          </router-link>
        </li>
        <li class="is-active">
          <router-link exact :to="{ name: 'newsletter-page', params: { id } }">
            {{ title }}
          </router-link>
        </li>
      </ul>
    </nav>

    <h1 class="title" v-text="title" />

    <progress
      v-if="!didLoad && isLoading"
      class="progress is-large is-warning"
      max="100"
    >
      Loading…
    </progress>

    <div v-if="didLoad">
      <template v-for="[prop] of Object.entries(page.frontmatter)">
        <p :key="prop">{{ prop }}</p>
      </template>
    </div>

    <div v-if="error" class="message is-danger">
      <div class="message-header">{{ error.name }}</div>
      <div class="message-body">
        <p class="content">{{ error.message }}</p>
        <div class="buttons">
          <button
            class="button is-danger has-text-weight-semibold"
            @click="reload"
          >
            Reload?
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
