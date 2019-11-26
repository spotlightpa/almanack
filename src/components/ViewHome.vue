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
  <div class="section container content">
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
    <div
      v-if="$auth.hasRole('editor') && !$api.loading"
      class="table-container"
    >
      <table
        class="table is-bordered is-striped is-narrow is-hoverable is-fullwidth"
      >
        <thead>
          <tr>
            <th>Slug</th>
            <th>Planned for date</th>
            <th>Published</th>
          </tr>
        </thead>
        <tbody>
          <template v-for="article of $api.contents">
            <tr :key="article.slug">
              <td>{{ article.slug }}</td>
              <td>
                {{
                  article.planning.scheduling.planned_publish_date | formatDate
                }}
              </td>
              <td>
                <a
                  v-if="article.workflow.status_code === 6"
                  :href="`https://www.inquirer.com${article.website_url}`"
                  target="_blank"
                  >Inquirer Link</a
                >
                <span v-else>No</span>
              </td>
            </tr>
            <tr :key="article.slug + '-details'">
              <td colspan="3">
                <details>
                  <summary>Read Budget Line</summary>
                  {{ article.planning.budget_line }}
                </details>
              </td>
            </tr>
          </template>
        </tbody>
      </table>
    </div>
  </div>
</template>
