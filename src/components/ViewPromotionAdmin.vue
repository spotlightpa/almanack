<script setup>
import { ref } from "vue";

import { get, post, listPromotions, postPromotion } from "@/api/client-v2.js";
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

// new promotion form
const isAdding = ref(false);
const newName = ref("");
const newDescription = ref("");
const newLink = ref("");
const newWidth = ref(0);
const newHeight = ref(0);
const newImageUrls = ref([]);

function startNew() {
  newName.value = "";
  newDescription.value = "";
  newLink.value = "";
  newWidth.value = 0;
  newHeight.value = 0;
  newImageUrls.value = [];
  isAdding.value = true;
}

const { apiState: saveState, exec: saveExec } = makeState();

async function saveNew() {
  await saveExec(() =>
    post(postPromotion, {
      id: null,
      name: newName.value,
      description: newDescription.value,
      link: newLink.value,
      width: newWidth.value,
      height: newHeight.value,
      image_urls: newImageUrls.value,
    })
  );
  if (!saveState.error) {
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
    <h1 class="title">Promotions</h1>

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
        <BulmaFieldInput v-model="newName" label="Name" />
        <BulmaFieldInput v-model="newDescription" label="Description" />
        <BulmaFieldInput v-model="newLink" label="Link URL" type="url" />
        <div class="is-flex mb-3" style="gap: 1rem">
          <BulmaFieldInput
            label="Width"
            inputmode="numeric"
            :model-value="newWidth || ''"
            @update:model-value="newWidth = +$event || 0"
          />
          <BulmaFieldInput
            label="Height"
            inputmode="numeric"
            :model-value="newHeight || ''"
            @update:model-value="newHeight = +$event || 0"
          />
        </div>
        <BulmaField
          label="Image URLs"
          help="One image URL per slot; one will be chosen randomly on each page load."
        >
          <SiteParamsFiles
            :files="newImageUrls"
            :file-props="fileList"
            @add="newImageUrls.push($event)"
            @remove="newImageUrls.splice($event, 1)"
          />
        </BulmaField>
        <ErrorSimple :error="saveState.error" />
        <div class="buttons">
          <button
            class="button is-success has-text-weight-semibold"
            :class="{ 'is-loading': saveState.isLoading }"
            type="button"
            @click="saveNew"
          >
            Save
          </button>
          <button
            class="button is-light has-text-weight-semibold"
            :disabled="saveState.isLoading || null"
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
        <div v-if="nextPage" class="buttons">
          <router-link
            :to="nextPage"
            class="button is-light has-text-weight-semibold"
          >
            Show More…
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
