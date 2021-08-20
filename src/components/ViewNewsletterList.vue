<script>
import { computed, toRefs, watch } from "@vue/composition-api";

import { makeState } from "@/api/service-util.js";
import { useClient } from "@/api/client.js";

import PageListItem from "@/api/spotlightpa-page-list-item.js";

import APILoader from "./APILoader.vue";
import PageListRow from "./PageListRow.vue";

export default {
  name: "ViewNewsletterList",
  components: { APILoader, PageListRow },
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
      ...toRefs(apiState),
      fetch,
      pages: PageListItem.from(apiState),
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
      link(id) {
        return {
          name: "newsletter-page",
          params: {
            id: "" + id,
          },
        };
      },
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
          <PageListRow
            v-for="page of pages"
            :key="page.id"
            :link="link(page.id)"
            :status="page.status"
            :label="page.internalID"
            :date="page.publishedAt"
            :hed="page.title"
            :dek="page.blurb"
            :image="page.image"
            :image-alt="page.image"
          />
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
