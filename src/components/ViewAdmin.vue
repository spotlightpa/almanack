<script setup>
import { ref } from "vue";
import { useRouter } from "vue-router";

import {
  get,
  post,
  listSharedArticles,
  postSharedArticleFromGDocs,
} from "@/api/client-v2.js";
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
  await gdocsExec(() =>
    post(postSharedArticleFromGDocs, {
      gdocs_id: id,
    })
  );
  if (gdocsState.error.value) {
    return;
  }

  // Trigger background image processing
  window.fetch("/api-background/images").catch(() => {});

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

    <LinkButtons label="News Partners">
      <LinkRoute
        label="External Editor View"
        to="shared-articles"
        :icon="['fas', 'file-invoice']"
      />
      <LinkRoute
        label="Preauthorization"
        to="domains"
        :icon="['fas', 'user-circle']"
      />
    </LinkButtons>

    <LinkButtons label="Spotlight PA promotions">
      <LinkRoute
        label="Homepage Editor"
        to="homepage-editor"
        :icon="['fas', 'newspaper']"
      />
      <LinkRoute
        label="Sidebar Items"
        to="sidebar-items"
        :icon="['fas', 'check-circle']"
      />
      <LinkRoute
        label="Sitewide Settings"
        to="site-params"
        :icon="['fas', 'sliders-h']"
      />
    </LinkButtons>

    <LinkButtons label="Spotlight PA pages">
      <LinkRoute
        label="Spotlight PA Articles"
        to="news-pages"
        :icon="['fas', 'file-signature']"
      />
      <LinkRoute
        label="Newsletter pages"
        to="newsletters"
        :icon="['fas', 'mail-bulk']"
      />
      <LinkRoute
        label="Election Features"
        to="election-features"
        :icon="['fas', 'check-circle']"
      />
    </LinkButtons>

    <LinkButtons label="Uploads">
      <LinkRoute
        label="Photo manager"
        to="image-uploader"
        :icon="['fa', 'file-image']"
      />
      <LinkRoute
        label="File manager"
        to="file-uploader"
        :icon="['fa', 'file-upload']"
      />
    </LinkButtons>
    <LinkButtons label="State College">
      <LinkRoute
        label="Articles"
        to="statecollege-pages"
        :icon="['fas', 'file-signature']"
      />
      <LinkRoute
        label="Frontpage Editor"
        to="state-college-editor"
        :icon="['fas', 'newspaper']"
      />
    </LinkButtons>

    <LinkButtons label="Tools">
      <button
        type="button"
        class="button is-primary is-small has-text-weight-semibold"
        @click="showComposer = !showComposer"
      >
        <span class="icon">
          <font-awesome-icon :icon="['fas', 'paper-plane']" />
        </span>
        <span v-text="!showComposer ? 'Compose Message' : 'Hide Message'" />
      </button>
      <LinkHref label="Embeds" href="https://www.spotlightpa.org/embeds/" />
      <LinkHref
        label="Legacy admin"
        href="https://www.spotlightpa.org/admin/"
      />
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
      />
    </keep-alive>

    <ErrorReloader class="mt-5" :error="apiState.error.value" @reload="fetch" />

    <h2 class="mt-5 title">Import Articles</h2>
    <LinkButtons>
      <LinkRoute
        label="Import from Arc"
        to="arc-import"
        :icon="['fas', 'pen-nib']"
      />
    </LinkButtons>

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
        />
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
      <ErrorSimple :error="gdocsState.error.value" />
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
                    <font-awesome-icon :icon="['far', 'newspaper']" />
                    {{ article.internalID }}
                  </router-link>
                  <TagDate
                    v-if="article.publicationDate"
                    :date="article.publicationDate"
                  />
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
                      />
                    </span>
                    <span v-text="article.statusVerbose"></span>
                  </span>

                  <router-link class="tag is-light" :to="article.detailsRoute">
                    <span class="icon">
                      <font-awesome-icon :icon="['fas', 'file-invoice']" />
                    </span>
                    <span>Partner view</span>
                  </router-link>

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
                  <a
                    v-if="article.isGdoc"
                    class="tag is-light"
                    :href="article.gdocsURL"
                    target="_blank"
                  >
                    <span class="icon is-size-6">
                      <font-awesome-icon :icon="['fas', 'link']" />
                    </span>
                    <span>Google Docs</span>
                  </a>
                  <router-link
                    v-if="article.pageRoute"
                    class="tag is-light"
                    :to="article.pageRoute"
                  >
                    <span class="icon">
                      <font-awesome-icon :icon="['fas', 'user-clock']" />
                    </span>
                    <span>Spotlight admin</span>
                  </router-link>
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

    <SpinnerProgress :is-loading="apiState.isLoadingThrottled.value" />
  </div>
</template>
