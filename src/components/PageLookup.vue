<script setup>
import { ref } from "vue";
import { useRouter } from "vue-router";

import { watchAPI } from "@/api/service-util.js";
import { get, getPage } from "@/api/client-v2.js";

const router = useRouter();
const searchBox = ref("");
const url = ref("");

async function lookup(fullurl) {
  if (!fullurl) return [null, null];

  const select = "-body";
  let path = null;
  try {
    path = new URL(fullurl).pathname;
  } catch (e) {
    return [null, e];
  }

  if (!path) return [null, null];

  return await get(getPage, { by: "urlpath", value: path, select });
}

const { apiState, computedObj } = watchAPI(() => url.value, lookup);

const invisible = computedObj((page) => {
  let name = page.file_path.match(/content\/(news|statecollege)\//)
    ? "news-page"
    : null;
  if (!name) {
    apiState.error.value = { message: "No admin associate with page." };
    return;
  }

  router.push({
    name,
    params: {
      id: "" + page.id,
    },
  });
});
</script>

<template>
  <form class="notification" @submit.prevent="url = searchBox">
    <label class="label">Get page by URL</label>
    <div class="field is-grouped">
      <p class="control is-expanded">
        <input
          v-model="searchBox"
          class="input"
          type="url"
          placeholder="https://www.spotlightpa.org/story/"
        />
      </p>
      <p class="control">
        <button
          class="button is-primary has-text-weight-semibold"
          :class="{ 'is-loading': apiState.isLoading.value }"
          @click="url = searchBox"
        >
          Search
        </button>
      </p>
      <p class="control">
        <BulmaPaste
          button-class="button is-success has-text-weight-semibold"
          @paste="
            url = $event;
            lookup();
          "
        ></BulmaPaste>
      </p>
    </div>
    <p
      v-if="apiState.error.value"
      class="help is-danger"
      v-text="apiState.error.value.message"
    ></p>
    {{ invisible }}
  </form>
</template>
