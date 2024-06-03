<script setup>
import {
  get,
  post,
  listArcByLastUpdated,
  postSharedArticleFromArc,
} from "@/api/client-v2.js";
import { makeState, watchAPI } from "@/api/service-util.js";
import SharedArticle from "@/api/shared-article.js";

const props = defineProps({
  page: String,
});
const { apiState, fetch, computedList, computedProp } = watchAPI(
  () => ({ page: props.page, refresh: props.page === "0" }),
  (params) => get(listArcByLastUpdated, params)
);

const { apiStateRefs: importState, exec: execImport } = makeState();

const articles = computedList("stories", (s) => SharedArticle.fromArc(s));
const nextPage = computedProp("next_page", (page) => ({
  name: "arc-import",
  query: { page },
}));

async function doImport(article) {
  await execImport(() =>
    post(postSharedArticleFromArc, { arc_id: article.sourceID })
  );
  await fetch();
}
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
  ></BulmaBreadcrumbs>

  <h2 class="title">Import from Arc</h2>

  <SpinnerProgress
    :is-loading="importState.isLoadingThrottled.value"
  ></SpinnerProgress>
  <ErrorSimple :error="importState.error.value"></ErrorSimple>

  <div class="table-container">
    <table class="table is-bordered is-striped is-narrow is-fullwidth">
      <tbody>
        <template v-for="article in articles" :key="article.sourceID">
          <tr>
            <td>
              <h3 class="mt-1 mb-1 title is-3">
                <router-link
                  v-if="article.id"
                  class="mr-2 middle"
                  :to="article.adminRoute"
                >
                  <font-awesome-icon
                    :icon="['far', 'newspaper']"
                  ></font-awesome-icon>
                  {{ article.arc.slug }}
                </router-link>
                <span v-else class="mr-2 middle">
                  <font-awesome-icon
                    :icon="['far', 'newspaper']"
                  ></font-awesome-icon>
                  {{ article.arc.slug }}
                </span>
                <TagDate :date="article.arc.plannedDate"></TagDate>
              </h3>

              <div class="mb-1 tags">
                <span class="tag is-small" :class="article.statusClass">
                  <span class="icon is-size-6">
                    <font-awesome-icon
                      :icon="
                        article.isShared
                          ? ['fas', 'check-circle']
                          : ['fas', 'pen-nib']
                      "
                    ></font-awesome-icon>
                  </span>
                  <span v-text="article.statusVerbose"></span>
                </span>
                <a
                  v-if="article.isArc"
                  class="tag is-light"
                  :href="article.arc.arcURL"
                  target="_blank"
                >
                  <span class="icon is-size-6">
                    <font-awesome-icon
                      :icon="['fas', 'link']"
                    ></font-awesome-icon>
                  </span>
                  <span>Arc</span>
                </a>
                <router-link
                  v-if="article.id"
                  class="tag is-light"
                  :to="article.detailsRoute"
                >
                  <span class="icon">
                    <font-awesome-icon
                      :icon="['fas', 'file-invoice']"
                    ></font-awesome-icon>
                  </span>
                  <span>Partner view</span>
                </router-link>
              </div>
              <p class="mb-2 content">
                {{ article.arc.budgetLine }}
              </p>
              <div v-if="!article.id" class="buttons">
                <button
                  class="button is-primary is-small has-text-weight-semibold"
                  type="button"
                  @click.prevent="doImport(article)"
                >
                  Import
                </button>
              </div>
            </td>
          </tr>
        </template>
      </tbody>
    </table>
  </div>

  <SpinnerProgress :is-loading="apiState.isLoading.value"></SpinnerProgress>
  <ErrorReloader :error="apiState.error.value" @reload="fetch"></ErrorReloader>

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
