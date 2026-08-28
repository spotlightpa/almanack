<script setup>
import { ref } from "vue";

import { get, post, listPromotions, postPromotion } from "@/api/client.ts";
import { makePromotion } from "@/api/promotion.ts";
import { makeState, watchAPI } from "@/api/service-util.js";
import { useFileList } from "@/api/file-list.js";

const props = defineProps({
  page: { default: "" },
});

const fileList = useFileList();

const { apiState, fetch, computedList, computedProp } = watchAPI(
  () => props.page,
  (page) => get(listPromotions, page ? { page } : undefined)
);

const promotions = computedList("promotions", (p) => p);

const nextPage = computedProp("next_page", (page) => ({
  name: "promotions",
  query: { page },
}));
const hasMore = computedProp("next_page", (v) => !!v);

// new promotion form
const isAdding = ref(false);
const newPromo = makePromotion();

function startNew() {
  newPromo.init();
  isAdding.value = true;
}

const { apiStateRefs: saveState, exec: saveExec } = makeState();

async function saveNew() {
  await saveExec(() => post(postPromotion, newPromo.toJSON()));
  if (!saveState.error.value) {
    isAdding.value = false;
    await fetch();
  }
}
</script>

<template>
  <MetaHead>
    <title>Promotions • Spotlight PA Almanack</title>
  </MetaHead>

  <div class="px-2">
    <BulmaBreadcrumbs
      :links="[
        { name: 'Admin', to: { name: 'admin' } },
        { name: 'Promotions', to: { name: 'promotions' } },
      ]"
    />
    <h1 class="title">Saved Promotion Sets</h1>

    <div class="mb-4">
      <button
        v-if="!isAdding"
        class="button is-primary has-text-weight-semibold"
        type="button"
        @click="startNew"
      >
        <span class="icon">
          <font-awesome-icon :icon="['fas', 'plus']" />
        </span>
        <span>New promotion set</span>
      </button>
      <div v-else class="box">
        <h2 class="title is-5">New promotion set</h2>
        <PromotionForm :promo="newPromo" :file-list="fileList" />
        <ErrorSimple :error="saveState.error.value" />
        <div class="buttons">
          <button
            class="button is-success has-text-weight-semibold"
            :class="{ 'is-loading': saveState.isLoadingThrottled.value }"
            type="button"
            @click="saveNew"
          >
            Save
          </button>
          <button
            class="button is-light has-text-weight-semibold"
            :disabled="saveState.isLoadingThrottled.value || null"
            type="button"
            @click="isAdding = false"
          >
            Cancel
          </button>
        </div>
      </div>
    </div>

    <APILoader
      :is-loading="apiState.isLoading.value"
      :reload="fetch"
      :error="apiState.error.value"
    >
      <div v-for="promo in promotions" :key="promo.id" class="zebra-row p-3">
        <PromotionListRow
          :model-value="promo"
          @update:model-value="fetch()"
          @delete="fetch()"
        />
      </div>

      <div class="zebra-row p-3">
        <p v-if="!promotions.length" class="has-text-grey">
          No promotion sets yet.
        </p>
        <div v-if="hasMore" class="buttons">
          <router-link
            :to="nextPage"
            class="button is-light has-text-weight-semibold"
          >
            Show More Promotions…
          </router-link>
        </div>
      </div>
    </APILoader>
  </div>
</template>

<style scoped>
.zebra-row {
  background-color: #fff;
}

.zebra-row:nth-child(even) {
  background-color: #fafafa;
}

.zebra-row + .zebra-row {
  border-top: 1px solid #dbdbdb;
}
</style>
