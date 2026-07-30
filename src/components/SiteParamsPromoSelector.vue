<script setup>
import { ref, watch } from "vue";

import { get, listPromotions } from "@/api/client-v2.js";
import { watchAPI } from "@/api/service-util.js";
import { debounce } from "@/utils/wait.ts";

defineEmits(["select"]);

const searchText = ref("");

// Debounced search text: only updates 400ms after the user stops typing.
// watchAPI watches this ref, so fetches are naturally debounced.
const debouncedSearch = ref("");
watch(
  searchText,
  debounce((val) => {
    debouncedSearch.value = val.trim();
  }, 400)
);

const { apiState } = watchAPI(
  () => debouncedSearch.value,
  (text) => get(listPromotions, text ? { text } : undefined)
);

const promotions = () => apiState.rawData.value?.promotions ?? [];
</script>

<template>
  <div class="promo-selector box">
    <h4 class="title is-6 mb-2">Copy from a promotion set</h4>

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
      v-if="!apiState.isLoading.value && promotions().length === 0"
      class="help has-text-grey"
    >
      {{
        debouncedSearch
          ? "No matching promotion sets found."
          : "No promotion sets yet."
      }}
    </p>

    <table
      v-if="promotions().length"
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
        <tr v-for="promo in promotions()" :key="promo.id">
          <td>{{ promo.name }}</td>
          <td class="has-text-grey is-size-7">{{ promo.description }}</td>
          <td class="is-narrow">
            {{ Array.isArray(promo.data) ? promo.data.length : 0 }}
          </td>
          <td class="is-narrow">
            <button
              type="button"
              class="button is-small is-link has-text-weight-semibold"
              @click="$emit('select', promo)"
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
