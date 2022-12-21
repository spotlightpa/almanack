<script>
import { ref } from "vue";

import {
  get,
  post,
  listArcByLastUpdated,
  postSharedArticleFromArc,
} from "@/api/client-v2.js";
import { makeState, watchAPI } from "@/api/service-util.js";
import SharedArticle from "@/api/shared-article.js";

import { formatDate } from "@/utils/time-format";

export default {
  props: ["page"],
  setup(props) {
    const first = ref(true);
    const { apiState, fetch, computer } = watchAPI(
      () => ({ page: props.page, refresh: props.page === "0" && !first.value }),
      (params) => get(listArcByLastUpdated, params)
    );

    const { apiStateRefs: importState, exec: execImport } = makeState();

    return {
      apiState,
      fetch,
      articles: computer((rawData) => {
        if (!rawData?.stories) {
          return [];
        }
        first.value = false;
        let { stories } = rawData;
        return stories.map((s) => SharedArticle.fromArc(s));
      }),
      nextPage: computer((rawData) => {
        let page = rawData?.next_page;
        if (!page) return null;
        return {
          name: "arc-import",
          query: { page },
        };
      }),

      importState,
      async doImport(article) {
        await execImport(() =>
          post(postSharedArticleFromArc, { arc_id: article.sourceID })
        );
        await fetch();
      },

      formatDate,
    };
  },
};
</script>

<template>
  <MetaHead>
    <title>Import from Arc • Spotlight PA Almanack</title>
  </MetaHead>
  <BulmaBreadcrumbs
    :links="[
      { name: 'Admin', to: { name: 'admin' } },
      {
        name: 'Import from Arc',
        to: { name: 'arc-import', query: { page: 0 } },
      },
    ]"
  />

  <h2 class="title">Import from Arc</h2>

  <SpinnerProgress :is-loading="importState.isLoadingThrottled.value" />
  <ErrorSimple :error="importState.error.value" />

  <div class="table-container">
    <table
      class="table is-bordered is-striped is-narrow is-hoverable is-fullwidth"
    >
      <thead>
        <tr>
          <th>Slug</th>
          <th>Status</th>
          <th>Links</th>
          <th>Last Updated</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="article in articles" :key="article.sourceID">
          <td>
            <span v-if="article.id">
              <font-awesome-icon :icon="['far', 'newspaper']" />
              {{ article.slug }}
            </span>
            <a v-else @click.prevent="doImport(article)">
              <font-awesome-icon :icon="['far', 'newspaper']" />
              {{ article.slug }}
            </a>
          </td>
          <td>
            <span class="tag is-small" :class="article.statusClass">
              <span class="icon is-size-6">
                <font-awesome-icon
                  :icon="
                    article.isShared
                      ? ['fas', 'check-circle']
                      : ['fas', 'pen-nib']
                  "
                />
              </span>
              <span v-text="article.statusVerbose"></span>
            </span>
          </td>
          <td>
            <span class="tags">
              <a
                v-if="article.isArc"
                class="tag is-light"
                :href="article.arc.arcURL"
                target="_blank"
              >
                <span class="icon is-size-6">
                  <font-awesome-icon :icon="['fas', 'link']" />
                </span>
                <span>Arc</span>
              </a>
              <router-link
                v-if="article.id"
                class="tag is-light"
                :to="article.detailsRoute"
              >
                <span class="icon">
                  <font-awesome-icon :icon="['fas', 'file-invoice']" />
                </span>
                <span>Partner view</span>
              </router-link>
            </span>
          </td>
          <td>
            {{ formatDate(article.lastUpdated) }}
          </td>
        </tr>
      </tbody>
    </table>
  </div>

  <SpinnerProgress :is-loading="apiState.isLoading.value" />
  <ErrorReloader :error="apiState.error.value" @reload="fetch" />

  <div class="buttons mt-5">
    <router-link
      v-if="nextPage"
      :to="nextPage"
      class="button is-primary has-text-weight-semibold"
    >
      Show Older Stories…
    </router-link>
  </div>
</template>
