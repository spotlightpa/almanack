<script>
import { reactive, ref, toRefs } from "@vue/composition-api";
import { useClient } from "@/api/hooks.js";
import { formatDate } from "@/utils/time-format.js";

export default {
  name: "AdminListPanel",
  props: {
    article: Object,
  },
  setup(props, { emit }) {
    let client = useClient();
    let apiStatus = reactive({
      error: null,
      isSavingNote: false,
      isMakingPlanned: false,
      isMakingAvailable: false,
      isRemoving: false,
      isRefreshing: false,
    });

    let showComposer = ref(false);

    async function updateArticle(ref) {
      apiStatus[ref] = true;
      [, apiStatus.error] = await client.saveArcArticle(props.article);
      if (apiStatus.error) {
        apiStatus[ref] = false;
        return;
      }
      emit("refresh", { apiStatus, ref });
    }

    return {
      ...toRefs(apiStatus),

      showComposer,
      formatDate,

      async saveNote() {
        await updateArticle("isSavingNote");
      },
      async makePlanned() {
        props.article.setStatusPlanned();
        await updateArticle("isMakingPlanned");
      },
      async makeAvailable() {
        props.article.setStatusAvailable();
        await updateArticle("isMakingAvailable");
      },
      async remove() {
        props.article.unsetStatus();
        await updateArticle("isRemoving");
      },
      async refreshArc() {
        props.article.setRefreshArc();
        await updateArticle("isRefreshing");
      },
    };
  },
};
</script>
<template>
  <div class="control">
    <h2 class="title is-spaced is-3">
      <ArticleSlugLine :article="article" />
    </h2>

    <p class="has-margin-top-negative">
      <strong>Byline:</strong>
      {{ article.byline }}
    </p>
    <p>
      <strong>Planned time:</strong>
      {{ formatDate(article.plannedDate) }}
    </p>
    <div class="field">
      <label class="label"> Publication Note: </label>
    </div>
    <div class="field has-addons">
      <div class="control is-expanded">
        <input v-model="article.note" class="input" />
      </div>
      <div class="control">
        <button
          class="button has-text-weight-semibold is-primary"
          :class="{ 'is-loading': isSavingNote }"
          @click="saveNote"
        >
          Save
        </button>
      </div>
    </div>
    <div v-if="error" class="message is-danger">
      <p class="message-header">{{ error.name }}</p>
      <p class="message-body">{{ error.message }}</p>
    </div>
    <div class="buttons has-margin-top-thin">
      <button
        v-if="!article.isPlanned || article.isAvailable"
        type="button"
        class="button is-warning has-text-weight-semibold"
        :class="{ 'is-loading': isMakingPlanned }"
        @click="makePlanned"
      >
        <span class="icon">
          <font-awesome-icon :icon="['fas', 'pen-nib']" />
        </span>
        <span> Show as Planned </span>
      </button>
      <button
        v-if="!article.isAvailable"
        type="button"
        class="button is-success has-text-weight-semibold"
        :class="{ 'is-loading': isMakingAvailable }"
        @click="makeAvailable"
      >
        <span class="icon">
          <font-awesome-icon :icon="['fas', 'check-circle']" />
        </span>
        <span> Make Available </span>
      </button>
      <button
        v-if="article.isPlanned || article.isAvailable"
        type="button"
        class="button is-light has-text-weight-semibold"
        :class="{ 'is-loading': isRemoving }"
        @click="remove"
      >
        <span class="icon">
          <font-awesome-icon :icon="['fas', 'trash-alt']" />
        </span>
        <span> Remove from Almanack </span>
      </button>
      <button
        type="button"
        class="button is-light has-text-weight-semibold"
        :class="{ 'is-loading': isRefreshing }"
        @click="refreshArc"
      >
        <span class="icon">
          <font-awesome-icon :icon="['fas', 'sync-alt']" />
        </span>
        <span> Refresh Arc Data </span>
      </button>

      <button
        type="button"
        class="button is-primary has-text-weight-semibold"
        @click="showComposer = !showComposer"
      >
        <span class="icon">
          <font-awesome-icon :icon="['fas', 'paper-plane']" />
        </span>
        <span v-text="!showComposer ? 'Compose Message' : 'Hide Message'" />
      </button>
    </div>
    <keep-alive>
      <EmailComposer
        v-if="showComposer"
        :initial-subject="article.emailSubject"
        :initial-body="article.emailBody"
        @hide="showComposer = false"
      />
    </keep-alive>
  </div>
</template>
