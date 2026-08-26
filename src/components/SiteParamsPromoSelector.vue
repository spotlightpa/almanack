<script setup>
import { ref, computed } from "vue";

import { get, listPromotions } from "@/api/client-v2.js";
import { watchAPI } from "@/api/service-util.js";
import { useDebouncedRef } from "@/utils/wait.ts";

const props = defineProps({
  filterWidth: { type: Number, default: 0 },
  filterHeight: { type: Number, default: 0 },
});

defineEmits(["select"]);

const searchText = ref("");
const debouncedSearch = useDebouncedRef(
  computed(() => searchText.value.trim()),
  400
);

const LIMIT = 20;

const { apiState, computedList } = watchAPI(
  () => debouncedSearch.value,
  (text) =>
    get(listPromotions, {
      text,
      width: props.filterWidth,
      height: props.filterHeight,
      limit: LIMIT + 1,
    })
);

const allResults = computedList("promotions", (p) => p);
const promotions = computed(() => allResults.value.slice(0, LIMIT));
const hasMore = computed(() => allResults.value.length > LIMIT);
</script>

<template>
  <div class="promo-selector box">
    <h4 class="title is-6 mb-2">Add from saved promotion set</h4>

    <div class="field">
      <div class="control" :class="{ 'is-loading': apiState.isLoading.value }">
        <input
          v-model="searchText"
          class="input is-small"
          type="search"
          placeholder="Search by name… (leave blank for recent)"
        />
      </div>
    </div>

    <p v-if="apiState.error.value" class="help is-danger">
      {{ apiState.error.value }}
    </p>

    <p
      v-if="!apiState.isLoading.value && promotions.length === 0"
      class="help has-text-grey"
    >
      {{
        debouncedSearch
          ? "No matching promotion sets found."
          : filterWidth || filterHeight
            ? `No promotion sets with size ${filterWidth}\xd7${filterHeight}.`
            : "No promotion sets yet."
      }}
    </p>

    <div
      v-for="promo in promotions"
      :key="promo.id"
      class="is-flex is-align-items-center is-justify-content-space-between py-2"
      style="border-bottom: 1px solid #dbdbdb"
    >
      <div>
        <p class="has-text-weight-semibold is-size-7">{{ promo.name }}</p>
        <p v-if="promo.description" class="has-text-grey is-size-7">
          {{ promo.description }}
        </p>
        <p class="has-text-grey is-size-7">
          {{ promo.image_urls?.length ?? 0 }} image{{
            promo.image_urls?.length !== 1 ? "s" : ""
          }}
        </p>
      </div>
      <button
        type="button"
        class="button is-small is-link has-text-weight-semibold"
        @click="$emit('select', promo)"
      >
        Use
      </button>
    </div>

    <p v-if="hasMore" class="help has-text-grey mt-2">
      More items not shown&hellip; refine your search to see them.
    </p>
  </div>
</template>

<style scoped>
.promo-selector {
  background: #f8f8f8;
  padding: 0.75rem;
}
</style>
