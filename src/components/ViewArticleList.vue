<script>
export default {
  created() {
    if (this.$auth.hasRole("editor")) {
      this.$api.load();
    } else {
      this.$api.loading = false;
    }
  }
};
</script>

<template>
  <div>
    <div class="content">
      <h2>
        Welcome, {{ $auth.user.user_metadata.full_name }}
        <small v-if="$auth.roles.length">
          (<span v-for="role of $auth.roles" :key="role" v-text="role"></span>)
        </small>
      </h2>
      <div v-if="!$auth.hasRole('editor')" class="message is-warning">
        <p class="message-body">
          You don't have permission to view upcoming articles, sorry. Please
          contact
          <a href="mailto:cjohnson@spotlightpa.org">cjohnson@spotlightpa.org</a>
          to request access.
        </p>
      </div>
    </div>
    <progress
      v-if="$api.loading"
      class="progress is-large is-warning"
      max="100"
    >
      Loading…
    </progress>
    <div v-if="$api.error" class="message is-danger ">
      <div class="message-body">
        <p>{{ $api.error }}</p>
        <div class="buttons">
          <button
            class="button is-danger has-text-weight-semibold"
            @click="$api.reload"
          >
            Reload?
          </button>
        </div>
      </div>
    </div>
    <nav v-if="$auth.hasRole('editor') && !$api.loading" class="panel is-black">
      <h2 class="panel-heading">
        Spotlight PA Articles
      </h2>
      <article
        v-for="article of $api.contents"
        :key="article.slug"
        class="panel-block"
      >
        <div class="control">
          <router-link to="" class="title is-3 has-text-link">
            <font-awesome-icon :icon="['far', 'newspaper']" />

            {{ article.slug }}
          </router-link>
          <p>
            <strong>Planned for:</strong>
            {{ article.planning.scheduling.planned_publish_date | formatDate }}

            <strong v-if="article.workflow.status_code === 6">
              <a
                :href="`https://www.inquirer.com${article.website_url}`"
                target="_blank"
                >Live Link
                <font-awesome-icon :icon="['fas', 'external-link-alt']" />
              </a>
            </strong>
          </p>
          <p class="has-margin-top-thin content is-small">
            {{ article.planning.budget_line }}
          </p>

          <div class="level is-mobile is-clipped">
            <div class="level-left">
              <p
                v-if="article.planning.story_length.word_count_planned"
                class="level-item"
              >
                <span>
                  <strong>Planned Word Count:</strong>
                  {{
                    article.planning.story_length.word_count_planned | intcomma
                  }}
                </span>
              </p>

              <p
                v-if="article.planning.story_length.word_count_actual"
                class="level-item is-hidden-mobile"
              >
                <span>
                  <strong>Word Count:</strong>
                  {{
                    article.planning.story_length.word_count_actual | intcomma
                  }}
                </span>
              </p>
              <p class="level-item is-hidden-mobile">
                <span>
                  <strong>Lines:</strong>
                  {{ article.planning.story_length.line_count_actual }}
                </span>
              </p>
              <p class="level-item is-hidden-mobile">
                <span>
                  <strong>Column inches:</strong>
                  {{ article.planning.story_length.inch_count_actual }}
                </span>
              </p>
            </div>
          </div>
        </div>
      </article>
    </nav>
  </div>
</template>
