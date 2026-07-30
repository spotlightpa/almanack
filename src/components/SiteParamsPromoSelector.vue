<script setup>
import { ref, watch } from "vue";

import { get, listPromotions } from "@/api/client-v2.js";

const emit = defineEmits(["select"]);

// ── recent promos (loaded once on mount) ─────────────────────────────────────
const recentPromos = ref([]);
const recentError = ref(null);
const recentLoading = ref(true);

async function loadRecent() {
  recentLoading.value = true;
  recentError.value = null;
  const [data, err] = await get(listPromotions);
  recentLoading.value = false;
  if (err) {
    recentError.value = err.message ?? String(err);
  } else {
    recentPromos.value = data?.promotions ?? [];
  }
}
loadRecent();

// ── search ───────────────────────────────────────────────────────────────────
const searchText = ref("");
const searchResults = ref([]);
const searchError = ref(null);
const searchLoading = ref(false);

let debounceTimer = null;

watch(searchText, (val) => {
  clearTimeout(debounceTimer);
  if (!val.trim()) {
    searchResults.value = [];
    searchError.value = null;
    searchLoading.value = false;
    return;
  }
  searchLoading.value = true;
  debounceTimer = setTimeout(() => doSearch(val.trim()), 400);
});

async function doSearch(text) {
  searchError.value = null;
  const [data, err] = await get(listPromotions, { text });
  searchLoading.value = false;
  if (err) {
    searchError.value = err.message ?? String(err);
  } else {
    searchResults.value = data?.promotions ?? [];
  }
}

// ── active list: search results when searching, else recent ──────────────────
function displayList() {
  return searchText.value.trim() ? searchResults.value : recentPromos.value;
}

function isLoading() {
  return searchText.value.trim() ? searchLoading.value : recentLoading.value;
}

function currentError() {
  return searchText.value.trim() ? searchError.value : recentError.value;
}

function selectPromo(promo) {
  emit("select", promo);
}
</script>

<template>
  <div class="promo-selector box">
    <h4 class="title is-6 mb-2">Copy from a promotion set</h4>

    <div class="field">
      <div class="control" :class="{ 'is-loading': isLoading() }">
        <input
          v-model="searchText"
          class="input is-small"
          type="search"
          placeholder="Search by name… (leave blank for recent)"
        />
      </div>
    </div>

    <p v-if="currentError()" class="help is-danger">{{ currentError() }}</p>

    <p
      v-if="!isLoading() && displayList().length === 0"
      class="help has-text-grey"
    >
      {{
        searchText.trim()
          ? "No matching promotion sets found."
          : "No promotion sets yet."
      }}
    </p>

    <table
      v-if="displayList().length"
      class="table is-narrow is-hoverable is-fullwidth is-bordered"
    >
      <thead>
        <tr>
          <th>Name</th>
          <th>Description</th>
          <th>Items</th>
          <th></th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="promo in displayList()" :key="promo.id">
          <td>{{ promo.name }}</td>
          <td class="has-text-grey is-size-7">{{ promo.description }}</td>
          <td class="is-narrow">
            {{ Array.isArray(promo.data) ? promo.data.length : 0 }}
          </td>
          <td class="is-narrow">
            <button
              type="button"
              class="button is-small is-link has-text-weight-semibold"
              @click="selectPromo(promo)"
            >
              Use
            </button>
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<style scoped>
.promo-selector {
  background: #f8f8f8;
  padding: 0.75rem;
}
</style>
