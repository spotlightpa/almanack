<script>
import { ref } from "@vue/composition-api";

import { useListAnyArc } from "@/api/hooks.js";

export default {
  name: "ViewAdmin",
  props: ["page"],
  metaInfo: {
    title: "Admin",
  },
  setup(props) {
    let {
      articles,
      didLoad,
      isLoading,
      isLoadingThrottled,
      load,
      nextPage,
      error,
    } = useListAnyArc(() => props.page);
    return {
      showBookmarklet: ref(false),
      showComposer: ref(false),
      didLoad,
      isLoading,
      isLoadingThrottled,
      load,
      error,
      articles,
      nextPage,
      async refresh({ apiStatus, ref }) {
        await load();
        apiStatus[ref] = false;
      },
    };
  },
};
</script>

<template>
  <div>
    <h1 class="title is-flex">
      Spotlight Administrator
      <template v-if="page">(page {{ page }})</template>
    </h1>

    <LinkButtons label="News Partners">
      <LinkButton
        label="External Editor View"
        to="articles"
        :icon="['fas', 'file-invoice']"
      />
      <LinkButton
        label="Preauthorization"
        to="domains"
        :icon="['fas', 'user-circle']"
      />
    </LinkButtons>

    <LinkButtons label="Spotlight PA promotions">
      <LinkButton
        label="Homepage Editor"
        to="homepage-editor"
        :icon="['fas', 'newspaper']"
      />
      <LinkButton
        label="Sidebar Items"
        to="sidebar-items"
        :icon="['fas', 'check-circle']"
      />
      <LinkButton
        label="Sitewide Settings"
        to="site-params"
        :icon="['fas', 'sliders-h']"
      />
    </LinkButtons>

    <LinkButtons label="Spotlight PA pages">
      <LinkButton
        label="Spotlight PA Articles"
        to="news-pages"
        :icon="['fas', 'file-signature']"
      />
      <LinkButton
        label="Newsletter pages"
        to="newsletters"
        :icon="['fas', 'mail-bulk']"
      />
    </LinkButtons>

    <LinkButtons label="Uploads">
      <LinkButton
        label="Photo manager"
        to="image-uploader"
        :icon="['fa', 'file-image']"
      />
      <LinkButton
        label="File manager"
        to="file-uploader"
        :icon="['fa', 'file-upload']"
      />
    </LinkButtons>
    <LinkButtons label="State College">
      <LinkButton
        label="Articles"
        to="statecollege-pages"
        :icon="['fas', 'file-signature']"
      />
      <LinkButton
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
      <LinkAButton label="Embeds" href="https://www.spotlightpa.org/embeds/" />
      <LinkAButton
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

    <ErrorReloader class="mt-5" :error="error" @reload="load" />

    <AdminList
      v-if="articles.length"
      class="mt-5"
      :articles="articles"
      title="Arc Articles"
      @refresh="refresh"
    />

    <div class="buttons mt-5">
      <router-link
        v-if="nextPage"
        :to="nextPage"
        class="button is-primary has-text-weight-semibold"
      >
        Show Older Stories…
      </router-link>
    </div>

    <SpinnerProgress :is-loading="isLoadingThrottled" />
  </div>
</template>
