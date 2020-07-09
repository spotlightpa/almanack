<script>
import ArticleList from "./ArticleList.vue";
import { useAuth, useListAvailableArc, useClient } from "@/api/hooks.js";

export default {
  name: "ViewArticleList",
  components: {
    ArticleList,
  },
  props: ["page"],
  metaInfo: {
    title: "Available Articles",
  },
  setup(props) {
    let { fullName, roles } = useAuth();
    let {
      articles,
      nextPage,
      didLoad,
      isLoading,
      load,
      error,
    } = useListAvailableArc(() => props.page);
    let client = useClient();

    return {
      articles,
      nextPage,
      didLoad,
      isLoading,
      load,
      error,
      fullName,
      roles,

      async redirectToSignup() {
        let [url, err] = await client.getSignupURL();
        if (err) {
          alert(`Something went wrong: ${err}`);
          return;
        }
        window.location = url;
      },
    };
  },
};
</script>

<template>
  <div>
    <h2 class="title">Welcome, {{ fullName }}</h2>
    <p class="content">
      Please note that this is an internal content distribution system, not
      intended for public use. Please
      <strong>do not share this URL</strong> with anyone besides the appointed
      contacts at your organization and please be mindful of the notes and
      embargos attached to each story. For assistance or if you have any
      questions, please contact Sarah Anne Hughes (<a
        href="mailto:shughes@spotlightpa.org"
        >shughes@spotlightpa.org</a
      >).
    </p>
    <p class="content">
      You should receive alerts about available stories from
      press@spotlightpa.org. Please add this address to your contacts to ensure
      that messages are not sent to spam. You can unsubscribe by following the
      unsubscribe link in the footer of an email, and you can
      <a href="#" @click.prevent="redirectToSignup">resubscribe here</a>.
    </p>

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

    <ArticleList
      v-if="articles.length"
      :articles="articles"
      title="Spotlight PA Articles"
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
