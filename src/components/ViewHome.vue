<script>
import Vue from "vue";
import FeedAPI from "./FeedAPI.vue";

export default {
  data() {
    return { feed: new Vue(FeedAPI) };
  }
};
</script>

<template>
  <div class="section container content">
    <h2>
      Hello {{ $auth.user.user_metadata.full_name }} (<span
        v-for="role of $auth.user.app_metadata.roles"
        :key="role"
        v-text="role"
      ></span
      >).
    </h2>
    <progress
      v-if="feed.loading"
      class="progress is-large is-warning"
      max="100"
    >
      Loading…
    </progress>

    <div v-if="feed.error" class="message is-danger ">
      <div class="message-body">
        <p>{{ feed.error }}</p>
        <div class="buttons">
          <button
            class="button is-danger has-text-weight-semibold"
            @click="feed.reload"
          >
            Reload?
          </button>
        </div>
      </div>
    </div>
    <div v-if="!feed.loading" class="table-container">
      <table
        class="table is-bordered is-striped is-narrow is-hoverable is-fullwidth"
      >
        <thead>
          <tr>
            <th>Slug</th>
            <th>Planned date</th>
            <th>Published</th>
          </tr>
        </thead>
        <tbody>
          <template v-for="article of feed.contents">
            <tr :key="article.slug">
              <td>{{ article.slug }}</td>
              <td>
                {{
                  article.planning.scheduling.planned_publish_date | formatDate
                }}
              </td>
              <td>
                {{ article.workflow.status_code === 6 ? "Y" : "N" }}
              </td>
            </tr>
            <tr :key="article.slug + '-details'">
              <td colspan="3">
                <details>
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
