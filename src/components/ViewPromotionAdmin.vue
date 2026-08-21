<script setup>
import { ref } from "vue";

import { get, post, listPromotions, postPromotion } from "@/api/client-v2.js";
import { makeState, watchAPI } from "@/api/service-util.js";
import { useFileList } from "@/api/file-list.js";

const fileList = useFileList();

const { apiState, fetch, computedList } = watchAPI(
  () => null,
  () => get(listPromotions)
);

const promotions = computedList("promotions", (p) => p);

function swap(event, i) {
  promotions.value[i] = event;
}

// new promotion form
const isAdding = ref(false);
const newName = ref("");
const newDescription = ref("");
const newWidth = ref(0);
const newHeight = ref(0);
const newItems = ref([]);

function startNew() {
  newName.value = "";
  newDescription.value = "";
  newWidth.value = 0;
  newHeight.value = 0;
  newItems.value = [];
  isAdding.value = true;
}

const { apiState: saveState, exec: saveExec } = makeState();

async function saveNew() {
  await saveExec(() =>
    post(postPromotion, {
      id: null,
      name: newName.value,
      description: newDescription.value,
      width: newWidth.value,
      height: newHeight.value,
      items: newItems.value,
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

    <APILoader
      :is-loading="apiState.isLoading.value"
      :reload="fetch"
      :error="apiState.error.value"
    >
      <div
        v-for="(promo, i) in promotions"
        :key="promo.id"
        class="zebra-row p-3"
      >
        <PromotionListRow
          :model-value="promo"
          @update:model-value="swap($event, i)"
        />
      </div>

      <div class="zebra-row p-3">
        <template v-if="!isAdding">
          <p v-if="!promotions.length" class="mb-3 has-text-grey">
            No promotion sets yet.
          </p>
          <button
            class="button is-primary has-text-weight-semibold"
            type="button"
            @click="startNew"
          >
            <span class="icon">
              <font-awesome-icon :icon="['fas', 'plus']" />
            </span>
            <span>New promotion set</span>
          </button>
        </template>

        <template v-else>
          <h2 class="title is-5">New promotion set</h2>
          <BulmaFieldInput
            v-model="newName"
            label="Name"
            placeholder="e.g. Rail sticky promo"
          />
          <BulmaFieldInput
            v-model="newDescription"
            label="Description"
            placeholder="Short description"
          />
          <div class="is-flex mb-3" style="gap: 1rem">
            <BulmaField v-slot="{ idForLabel }" label="Width">
              <input
                :id="idForLabel"
                :value="newWidth || ''"
                class="input"
                inputmode="numeric"
                @change="newWidth = +$event.target.value || 0"
              />
            </BulmaField>
            <BulmaField v-slot="{ idForLabel }" label="Height">
              <input
                :id="idForLabel"
                :value="newHeight || ''"
                class="input"
                inputmode="numeric"
                @change="newHeight = +$event.target.value || 0"
              />
            </BulmaField>
          </div>
          <BulmaField
            label="Image URLs"
            help="One image URL per slot; one will be chosen randomly on each page load."
          >
            <SiteParamsFiles
              :files="newItems"
              :file-props="fileList"
              @add="newItems.push($event)"
              @remove="newItems.splice($event, 1)"
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
        </template>
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
