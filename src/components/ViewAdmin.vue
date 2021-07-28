<script>
import { ref } from "@vue/composition-api";

import AdminList from "./AdminList.vue";
import EmailComposer from "./EmailComposer.vue";

import { useListAnyArc } from "@/api/hooks.js";

import ImageUploader from "./ImageUploader.vue";

export default {
  name: "ViewAdmin",
  components: {
    AdminList,
    EmailComposer,
    ImageUploader,
  },
  props: ["page"],
  metaInfo: {
    title: "Admin",
  },
  setup(props) {
    let { articles, didLoad, isLoading, load, nextPage, error } = useListAnyArc(
      () => props.page
    );

    return {
      showComposer: ref(false),

      didLoad,
      isLoading,
      load,
      error,
      articles,
      nextPage,

      async refresh({ apiStatus, ref }) {
        await load();
        apiStatus[ref] = false;
      },

      navLinks: [
        {
          to: "articles",
          icon: ["fas", "file-invoice"],
          text: "External Editor View",
        },
        {
          to: "editors-picks",
          icon: ["fas", "newspaper"],
          text: "Homepage Editor",
        },
        {
          to: "spotlightpa-articles",
          icon: ["fas", "file-signature"],
          text: "Spotlight PA Articles",
        },
        {
          to: "newsletters",
          icon: ["fas", "mail-bulk"],
          text: "Newsletter pages",
        },

        {
          to: "uploader",
          icon: ["fa", "file-image"],
          text: "Photo manager",
        },
        {
          to: "file-uploader",
          icon: ["fa", "file-upload"],
          text: "File manager",
        },
        {
          to: "domains",
          icon: ["fas", "user-circle"],
          text: "Preauthorization",
        },
      ],
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

    <nav class="buttons">
      <router-link
        v-for="(link, i) in navLinks"
        :key="i"
        :to="{ name: link.to }"
        class="button is-small is-light has-text-weight-semibold"
      >
        <span class="icon">
          <font-awesome-icon :icon="link.icon" />
        </span>
        <span v-text="link.text" />
      </router-link>
    </nav>

    <details class="content">
      <summary>Tools</summary>
      <p>
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
        bookmarklet
        <sup
          ><a
            title="What's a bookmarklet?"
            href="https://support.mozilla.org/en-US/kb/bookmarklets-perform-common-web-page-tasks"
            >?</a
          ></sup
        >
      </p>
    </details>

    <div class="level">
      <div class="level-left">
        <div class="level-item">
          <div class="control">
            <label class="label">Upload an image</label>
            <ImageUploader />
          </div>
        </div>
        <div class="level-item">
          <div class="control">
            <label class="label">Compose a message</label>
            <button
              type="button"
              class="button is-primary has-text-weight-semibold"
              @click="showComposer = !showComposer"
            >
              <span class="icon">
                <font-awesome-icon :icon="['fas', 'paper-plane']" />
              </span>
              <span
                v-text="!showComposer ? 'Compose Message' : 'Hide Message'"
              />
            </button>
          </div>
        </div>
      </div>
    </div>

    <keep-alive>
      <EmailComposer
        v-if="showComposer"
        initial-subject="Subject"
        initial-body="Email body"
        @hide="showComposer = false"
      />
    </keep-alive>

    <progress
      v-if="!didLoad && isLoading"
      class="progress is-large is-warning"
      max="100"
    >
      Loading…
    </progress>

    <div v-if="error" class="message is-danger">
      <div class="message-header">{{ error.name }}</div>
      <div class="message-body">
        <p class="content">{{ error.message }}</p>
        <div class="buttons">
          <button
            class="button is-danger has-text-weight-semibold"
            @click="load"
          >
            Reload?
          </button>
        </div>
      </div>
    </div>

    <AdminList
      v-if="articles.length"
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
  </div>
</template>
