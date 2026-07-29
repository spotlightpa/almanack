<script setup>
import { ref, reactive, computed } from "vue";

import { get, post, listPromotions, postPromotion } from "@/api/client-v2.js";
import { makeState } from "@/api/service-util.js";

// ── list state ──────────────────────────────────────────────────────────────
const { apiState: listState, exec: listExec } = makeState();
const promotions = computed(() => listState.rawData?.promotions ?? []);

async function reload() {
  await listExec(() => get(listPromotions));
}
reload();

// ── save state ───────────────────────────────────────────────────────────────
const { apiState: saveState, exec: saveExec } = makeState();

// ── editing ──────────────────────────────────────────────────────────────────
const editing = ref(null); // the promotion object currently being edited
const dataError = ref("");

function blankPromotion() {
  return {
    id: 0,
    name: "",
    description: "",
    width: 0,
    height: 0,
    data: "{}",
  };
}

function startEdit(promo) {
  editing.value = reactive({
    ...promo,
    data: JSON.stringify(promo.data ?? {}, null, 2),
  });
  dataError.value = "";
}

function startNew() {
  editing.value = reactive(blankPromotion());
  dataError.value = "";
}

function cancelEdit() {
  editing.value = null;
  dataError.value = "";
}

async function save() {
  let parsedData;
  try {
    parsedData = JSON.parse(editing.value.data);
    dataError.value = "";
  } catch (e) {
    dataError.value = "Data must be valid JSON: " + e.message;
    return;
  }

  const payload = {
    id: editing.value.id,
    name: editing.value.name,
    description: editing.value.description,
    width: Number(editing.value.width),
    height: Number(editing.value.height),
    data: parsedData,
  };

  await saveExec(() => post(postPromotion, payload));
  if (!saveState.error) {
    editing.value = null;
    await reload();
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

    <!-- ── editor panel ───────────────────────────────────────────── -->
    <div v-if="editing" class="box mb-5">
      <h2 class="subtitle mb-4">
        {{ editing.id ? `Edit promotion #${editing.id}` : "New promotion" }}
      </h2>

      <BulmaFieldInput
        v-model="editing.name"
        label="Name"
        placeholder="Promo name"
      />

      <BulmaFieldInput
        v-model="editing.description"
        label="Description"
        placeholder="Short description"
      />

      <div class="columns">
        <div class="column">
          <BulmaFieldInput
            v-model.number="editing.width"
            label="Width"
            type="number"
            placeholder="0"
          />
        </div>
        <div class="column">
          <BulmaFieldInput
            v-model.number="editing.height"
            label="Height"
            type="number"
            placeholder="0"
          />
        </div>
      </div>

      <div class="field">
        <label class="label">Data (JSON)</label>
        <div class="control">
          <textarea
            v-model="editing.data"
            class="textarea is-family-monospace"
            rows="6"
            spellcheck="false"
          />
        </div>
        <p v-if="dataError" class="help is-danger">{{ dataError }}</p>
      </div>

      <div class="buttons">
        <button
          class="button is-primary has-text-weight-semibold"
          :class="{ 'is-loading': saveState.isLoading }"
          :disabled="saveState.isLoading || null"
          @click="save"
        >
          Save
        </button>
        <button
          class="button is-light has-text-weight-semibold"
          :disabled="saveState.isLoading || null"
          @click="cancelEdit"
        >
          Cancel
        </button>
      </div>

      <ErrorSimple :error="saveState.error" />
    </div>

    <!-- ── toolbar ────────────────────────────────────────────────── -->
    <div v-if="!editing" class="mb-4 buttons">
      <button
        class="button is-primary has-text-weight-semibold"
        @click="startNew"
      >
        <span class="icon">
          <font-awesome-icon :icon="['fas', 'plus']" />
        </span>
        <span>New promotion</span>
      </button>
    </div>

    <!-- ── list ───────────────────────────────────────────────────── -->
    <APILoader
      :is-loading="listState.isLoading"
      :reload="reload"
      :error="listState.error"
    >
      <p v-if="!promotions.length" class="has-text-grey">
        No promotions yet.
      </p>

      <table v-else class="table is-bordered is-striped is-narrow is-fullwidth">
        <thead>
          <tr>
            <th>ID</th>
            <th>Name</th>
            <th>Description</th>
            <th>Width</th>
            <th>Height</th>
            <th>Updated</th>
            <th></th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="promo in promotions" :key="promo.id">
            <td class="is-narrow has-text-grey">#{{ promo.id }}</td>
            <td>{{ promo.name }}</td>
            <td>{{ promo.description }}</td>
            <td class="is-narrow">{{ promo.width || "—" }}</td>
            <td class="is-narrow">{{ promo.height || "—" }}</td>
            <td class="is-narrow is-size-7 has-text-grey">
              {{ new Date(promo.updated_at).toLocaleString() }}
            </td>
            <td class="is-narrow">
              <button
                class="button is-small is-info has-text-weight-semibold"
                @click="startEdit(promo)"
              >
                Edit
              </button>
            </td>
          </tr>
        </tbody>
      </table>
    </APILoader>
  </div>
</template>
