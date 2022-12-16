<script>
import { ref } from "vue";

import { get, listSharedArticles } from "@/api/client-v2.js";
import { watchAPI } from "@/api/service-util.js";
import SharedArticle from "@/api/shared-article.js";

export default {
  name: "ViewAdmin",
  props: ["page"],
  setup(props) {
    const { apiState, fetch, computer } = watchAPI(
      () => props.page || 0,
      (page) => get(listSharedArticles, { page, show: "all", select: "-body" })
    );

    return {
      showBookmarklet: ref(false),
      showComposer: ref(false),

      apiState,
      fetch,
      articles: computer((rawData) =>
        (rawData?.stories ?? []).map((a) => new SharedArticle(a))
      ),
      nextPage: computer((rawData) => {
        let page = rawData?.next_page;
        if (!page) return null;
        return {
          name: "admin",
          query: { page },
        };
      }),
    };
  },
};
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
        to="articles"
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

    <li v-for="article in articles" :key="article.id">
      {{ article.id }}
    </li>

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
