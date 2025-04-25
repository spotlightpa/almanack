<script setup>
import { ref } from "vue";
import { useRouter } from "vue-router";

import {
  get,
  post,
  listSharedArticles,
  postSharedArticleFromGDocs,
} from "@/api/client-v2.js";
import { processGDocsDoc } from "@/api/gdocs.js";
import { makeState, watchAPI } from "@/api/service-util.js";
import SharedArticle from "@/api/shared-article.js";

const props = defineProps({
  page: String,
});
const router = useRouter();

const { apiState, fetch, computedList, computedProp } = watchAPI(
  () => props.page || 0,
  (page) => get(listSharedArticles, { page, show: "all" })
);

const gdocsImportURL = ref("");
const { apiStateRefs: gdocsState, exec: gdocsExec } = makeState();
async function importGDocsURL(id) {
  await gdocsExec(async () => {
    let [, err] = await processGDocsDoc(id);
    if (err) {
      return [null, err];
    }
    return await post(postSharedArticleFromGDocs, {
      external_gdocs_id: id,
      force_update: false,
    });
  });
  if (gdocsState.error.value) {
    return;
  }

  let article = new SharedArticle(gdocsState.rawData.value);
  router.push({
    name: "shared-article-admin",
    params: {
      id: "" + article.id,
    },
  });
}

const showBookmarklet = ref(false);
const showComposer = ref(false);

const articles = computedList("stories", (a) => new SharedArticle(a));
const nextPage = computedProp("next_page", (page) => ({
  name: "admin",
  query: { page },
}));
</script>

<template>
  <MetaHead>
    <title>Admin • Spotlight PA</title>
  </MetaHead>
  <div>
    <h1 class="title is-flex">
      Spotlight Administrator
      <template v-if="page">(page {{ page }})</template>
    </h1>

    <LinkButtons label="News partners">
      <LinkRoute
        label="Partner Article List"
        to="shared-articles"
        :icon="['fas', 'file-invoice']"
      ></LinkRoute>
      <LinkRoute
        label="Preauthorization"
        to="domains"
        :icon="['fas', 'user-circle']"
      ></LinkRoute>
    </LinkButtons>

    <LinkButtons label="Spotlight PA promotions">
      <LinkRoute
        label="Sitewide Settings"
        to="site-params"
        :icon="['fas', 'sliders-h']"
      ></LinkRoute>
      <LinkRoute
        label="Sidebar Items"
        to="sidebar-items"
        :icon="['fas', 'check-circle']"
      ></LinkRoute>
      <LinkRoute
        label="Donor Walls"
        to="donor-wall"
        :icon="['fas', 'receipt']"
      ></LinkRoute>
    </LinkButtons>

    <LinkButtons label="Landing pages">
      <LinkRoute
        label="Homepage Editor"
        to="homepage-editor"
        :icon="['fas', 'newspaper']"
      ></LinkRoute>
      <LinkRoute
        label="State College Frontpage"
        to="state-college-editor"
        :icon="['fas', 'newspaper']"
      ></LinkRoute>
    </LinkButtons>

    <LinkButtons label="Spotlight PA article pages">
      <LinkRoute
        label="News Articles"
        to="news-pages"
        :icon="['fas', 'file-signature']"
      ></LinkRoute>
      <LinkRoute
        label="State College Articles"
        to="statecollege-pages"
        :icon="['fas', 'file-signature']"
      ></LinkRoute>
      <LinkRoute
        label="Berks County Articles"
        to="berks-pages"
        :icon="['fas', 'file-signature']"
      ></LinkRoute>
    </LinkButtons>

    <LinkButtons label="Uploads">
      <LinkRoute
        label="Photo manager"
        to="image-uploader"
        :icon="['fa', 'file-image']"
      ></LinkRoute>
      <LinkRoute
        label="File manager"
        to="file-uploader"
        :icon="['fa', 'file-upload']"
      ></LinkRoute>
    </LinkButtons>

    <LinkButtons label="Tools">
      <button
        type="button"
        class="button is-primary is-small has-text-weight-semibold"
        @click="showComposer = !showComposer"
      >
        <span class="icon">
          <font-awesome-icon :icon="['fas', 'paper-plane']"></font-awesome-icon>
        </span>
        <span
          v-text="!showComposer ? 'Compose Message' : 'Hide Message'"
        ></span>
      </button>
      <LinkHref
        label="Embeds"
        href="https://www.spotlightpa.org/embeds/"
      ></LinkHref>
      <LinkHref
        label="Legacy admin"
        href="https://www.spotlightpa.org/admin/"
      ></LinkHref>
      <LinkHref
        label="Set up Gmail signature"
        href="https://gmailsig.spotlightpa.org/"
      ></LinkHref>
      <button
        type="button"
        class="button is-light is-small has-text-weight-semibold"
        @click="showBookmarklet = !showBookmarklet"
      >
        Bookmarklet
      </button>
    </LinkButtons>

    <template v-if="showBookmarklet">
      <p>
        Bookmarklet:
        <a
          href="javascript:
(()=>{
  let match = location.href.match(/\/\d{4}\/\d\d\/([\w-.]+)\/?$/);
  if (!match) {
    alert('Not on Spotlight PA article');
    return;
  }
  let [, slug] = match;
  window.location = 'https://almanack.data.spotlightpa.org/api/bookmarklet/' + slug;
})();
        "
        >
          Jump to admin
        </a>
      </p>
      <p>
        <a
          href="https://support.mozilla.org/en-US/kb/bookmarklets-perform-common-web-page-tasks"
          >What's a bookmarklet?</a
        >
      </p>
    </template>

    <keep-alive>
      <EmailComposer
        v-if="showComposer"
        class="mt-5"
        initial-subject="Subject"
        initial-body="Email body"
        @hide="showComposer = false"
      ></EmailComposer>
    </keep-alive>

    <ErrorReloader
      class="mt-5"
      :error="apiState.error.value"
      @reload="fetch"
    ></ErrorReloader>

    <h2 class="mt-5 title">Import Articles</h2>

    <label for="gdocs-importer" class="mt-4 label">
      Import from Google Docs
    </label>
    <form
      class="field is-grouped"
      @submit.prevent="importGDocsURL(gdocsImportURL)"
    >
      <div class="control is-expanded">
        <input
          id="gdocs-importer"
          v-model="gdocsImportURL"
          class="input"
          placeholder="https://docs.google.com/document/d/abc123/edit"
        />
      </div>
      <div class="control">
        <BulmaPaste
          @paste="
            gdocsImportURL = $event;
            importGDocsURL(gdocsImportURL);
          "
        ></BulmaPaste>
      </div>
      <div class="control">
        <button
          class="button is-success has-text-weight-semibold"
          :class="gdocsState.isLoading.value && 'is-loading'"
          :disabled="!gdocsImportURL || null"
        >
          Import
        </button>
      </div>
    </form>
    <div class="field">
      <p class="help">Document must be shared with Spotlight PA.</p>
    </div>
    <div v-if="gdocsState.error.value" class="field">
      <ErrorSimple :error="gdocsState.error.value"></ErrorSimple>
    </div>

    <h2 class="mt-5 title">Shareable Articles</h2>

    <div class="table-container">
      <table class="table is-bordered is-striped is-narrow is-fullwidth">
        <tbody>
          <template v-for="article in articles" :key="article.id">
            <tr>
              <td>
                <h3 class="mt-1 mb-1 title is-3">
                  <router-link class="mr-2 middle" :to="article.adminRoute">
                    <font-awesome-icon
                      :icon="['far', 'newspaper']"
                    ></font-awesome-icon>
                    {{ article.internalID }}
                  </router-link>
                  <TagDate
                    v-if="article.publicationDate"
                    :date="article.publicationDate"
                  ></TagDate>
                </h3>
                <div class="mb-1 tags">
                  <span
                    class="tag is-small has-text-weight-semibold"
                    :class="article.statusClass"
                  >
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
                  <TagLink
                    :to="article.detailsRoute"
                    :icon="['fas', 'file-invoice']"
                  >
                    Partner view
                  </TagLink>
                  <TagLink
                    v-if="article.isArc"
                    :href="article.arc.arcURL"
                    :icon="['fas', 'link']"
                  >
                    Arc
                  </TagLink>
                  <TagLink
                    v-if="article.isGDoc"
                    :href="article.gdocsURL"
                    :icon="['fas', 'link']"
                  >
                    Google Docs
                  </TagLink>
                  <TagLink
                    v-if="article.pageRoute"
                    :to="article.pageRoute"
                    :icon="['fas', 'user-clock']"
                  >
                    Spotlight admin
                  </TagLink>
                </div>
                <p class="mb-1 content">
                  {{ article.budget }}
                </p>
              </td>
            </tr>
          </template>
        </tbody>
      </table>
    </div>

    <div class="buttons mt-5">
      <router-link
        v-if="nextPage"
        :to="nextPage"
        class="button is-primary has-text-weight-semibold"
      >
        Show Older Stories…
      </router-link>
    </div>

    <SpinnerProgress
      :is-loading="apiState.isLoadingThrottled.value"
    ></SpinnerProgress>
  </div>
</template>
