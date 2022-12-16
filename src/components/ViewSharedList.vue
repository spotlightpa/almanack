<script>
import { useAuth } from "@/api/hooks.js";
import { get, getSignupURL, listSharedArticles } from "@/api/client-v2.js";
import { watchAPI } from "@/api/service-util.js";
import SharedArticle from "@/api/shared-article.js";

export default {
  props: ["page"],
  setup(props) {
    let { fullName, roles } = useAuth();

    const { apiState, fetch, computer } = watchAPI(
      () => props.page || 0,
      (page) => get(listSharedArticles, { page })
    );

    return {
      apiState,
      fetch,
      articles: computer((rawData) =>
        (rawData?.stories ?? []).map((a) => new SharedArticle(a))
      ),
      nextPage: computer((rawData) => {
        let page = rawData?.next_page;
        if (!page) return null;
        return {
          name: "articles",
          query: { page },
        };
      }),

      fullName,
      roles,

      async redirectToSignup() {
        let [url, err] = await get(getSignupURL);
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
  <MetaHead>
    <title>Shared Articles • Spotlight PA Almanack</title>
  </MetaHead>

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

    <SpinnerProgress :is-loading="apiState.isLoading.value" />
    <ErrorReloader :error="apiState.error.value" @reload="fetch" />

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
