<script>
import ArticleList from "./ArticleList.vue";
import APILoader from "./APILoader.vue";
import { useAuth, useListAvailableArc, useClient } from "@/api/hooks.js";

export default {
  name: "ViewArticleList",
  components: {
    ArticleList,
    APILoader,
  },
  metaInfo: {
    title: "Available Articles",
  },
  setup() {
    let { fullName, roles } = useAuth();
    let { articles, isLoading, load, error } = useListAvailableArc();
    let client = useClient();

    return {
      isLoading,
      load,
      error,
      fullName,
      roles,
      articles,

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
    <h2 class="title">
      Welcome, {{ fullName }}
      <small v-if="roles.length > 1"> ({{ roles | commaand }}) </small>
      <small v-if="roles.length === 0"> (Not Authorized) </small>
    </h2>
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

    <APILoader :is-loading="isLoading" :reload="load" :error="error">
      <ArticleList
        v-if="articles.length"
        :articles="articles"
        title="Spotlight PA Articles"
      ></ArticleList>
    </APILoader>
  </div>
</template>
