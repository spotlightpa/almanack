<script setup>
import { ref } from "vue";

import { get, post, listPromotions, postPromotion } from "@/api/client-v2.js";
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

// new promotion form
const isAdding = ref(false);
const newPromo = makePromotion();

function startNew() {
  newPromo.init();
  isAdding.value = true;
}

const { apiState: saveState, exec: saveExec } = makeState();

async function saveNew() {
  await saveExec(() => post(postPromotion, newPromo.toJSON(null)));
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
        <BulmaFieldInput v-model="newPromo.name" label="Name" />
        <BulmaTextarea
          v-model="newPromo.description"
          label="Description"
          placeholder="Short description"
          :rows="2"
        />
        <div class="is-flex mb-3" style="gap: 1rem">
          <BulmaFieldInput
            label="Width"
            inputmode="numeric"
            :model-value="newPromo.width || ''"
            @update:model-value="newPromo.width = +$event || 0"
          />
          <BulmaFieldInput
            label="Height"
            inputmode="numeric"
            :model-value="newPromo.height || ''"
            @update:model-value="newPromo.height = +$event || 0"
          />
        </div>
        <BulmaFieldInput
          v-model="newPromo.link"
          label="Link URL"
          type="url"
          placeholder="https://www.spotlightpa.org/donate/"
        />
        <BulmaFieldInput
          v-model="newPromo.bannerLabel"
          label="Banner label"
          placeholder="Sponsored by Acme"
          help="Text accompanying a banner specifying the sponsor or presenter"
        />
        <BulmaFieldInput
          v-model="newPromo.bannerLabelLink"
          label="Banner label link"
          type="url"
          placeholder="https://www.spotlightpa.org/support/"
          help="Link that clicking the ad label will go to"
        />
        <BulmaTextarea
          v-model="newPromo.imageDescription"
          label="Image description (alt text)"
          help="For blind readers and search engines"
        />
        <BulmaField
          label="Images"
          help="If multiple images are provided for the same promotion, each page load will select one randomly"
        >
          <SiteParamsFiles
            :files="newPromo.imageUrls"
            :file-props="fileList"
            @add="newPromo.imageUrls.push($event)"
            @remove="newPromo.imageUrls.splice($event, 1)"
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
