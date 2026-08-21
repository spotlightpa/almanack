<script setup>
import { ref, reactive } from "vue";

import { get, post, listPromotions, postPromotion } from "@/api/client-v2.js";
import { makeState, watchAPI } from "@/api/service-util.js";
import { useFileList } from "@/api/file-list.js";

// ── file list (for image picker) ─────────────────────────────────────────────
const fileList = useFileList();

// ── list state ───────────────────────────────────────────────────────────────
const { apiState, fetch, computedList } = watchAPI(
  () => null,
  () => get(listPromotions)
);

const promotions = computedList("promotions", (p) => p);

// ── save state ───────────────────────────────────────────────────────────────
const { apiState: saveState, exec: saveExec } = makeState();

// ── editing ──────────────────────────────────────────────────────────────────
const editing = ref(null);

function startEdit(promo) {
  editing.value = reactive({
    id: promo.id,
    name: promo.name,
    description: promo.description,
    width: promo.width,
    height: promo.height,
    items: [...(promo.items ?? [])],
  });
}

function startNew() {
  editing.value = reactive({
    id: 0,
    name: "",
    description: "",
    width: 0,
    height: 0,
    items: [],
  });
}

function cancelEdit() {
  editing.value = null;
}

async function save() {
  await saveExec(() => post(postPromotion, editing.value));
  if (!saveState.error) {
    editing.value = null;
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
      <div class="mb-4 buttons">
        <button
          class="button is-primary has-text-weight-semibold"
          @click="startNew"
        >
          <span class="icon">
            <font-awesome-icon :icon="['fas', 'plus']" />
          </span>
          <span>New promotion set</span>
        </button>
      </div>

      <p v-if="!promotions.length" class="has-text-grey">
        No promotion sets yet.
      </p>

      <div v-for="promo in promotions" :key="promo.id" class="zebra-row p-3">
        <template v-if="editing && editing.id === promo.id">
          <!-- ── inline editor ──────────────────────────────────── -->
          <BulmaFieldInput
            v-model="editing.name"
            label="Name"
            placeholder="e.g. Rail sticky promo"
          />
          <BulmaFieldInput
            v-model="editing.description"
            label="Description"
            placeholder="Short description"
          />
          <div class="is-flex mb-3" style="gap: 1rem">
            <BulmaField v-slot="{ idForLabel }" label="Width">
              <input
                :id="idForLabel"
                v-model.number="editing.width"
                class="input"
                inputmode="numeric"
              />
            </BulmaField>
            <BulmaField v-slot="{ idForLabel }" label="Height">
              <input
                :id="idForLabel"
                v-model.number="editing.height"
                class="input"
                inputmode="numeric"
              />
            </BulmaField>
          </div>
          <BulmaField
            label="Image URLs"
            help="One image URL per slot; one will be chosen randomly on each page load."
          >
            <SiteParamsFiles
              :files="editing.items"
              :file-props="fileList"
              @add="editing.items.push($event)"
              @remove="editing.items.splice($event, 1)"
            />
          </BulmaField>
          <div class="mt-3 buttons">
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
        </template>

        <template v-else>
          <!-- ── list row ───────────────────────────────────────── -->
          <div
            class="is-flex is-justify-content-space-between is-align-items-start"
          >
            <div>
              <p class="has-text-weight-semibold">
                {{ promo.name }}
                <span class="has-text-grey is-size-7"> #{{ promo.id }}</span>
              </p>
              <p v-if="promo.description" class="is-size-7 has-text-grey">
                {{ promo.description }}
              </p>
              <p class="is-size-7 has-text-grey">
                {{ promo.width }}×{{ promo.height }}px &middot;
                {{ promo.items?.length ?? 0 }} image{{
                  promo.items?.length !== 1 ? "s" : ""
                }}
                &middot;
                {{ new Date(promo.updated_at).toLocaleString() }}
              </p>
            </div>
            <button
              class="button is-small is-info has-text-weight-semibold"
              @click="startEdit(promo)"
            >
              Edit
            </button>
          </div>
        </template>
      </div>

      <!-- ── new promotion form (when no id yet) ────────────────── -->
      <div v-if="editing && editing.id === 0" class="zebra-row p-3">
        <BulmaFieldInput
          v-model="editing.name"
          label="Name"
          placeholder="e.g. Rail sticky promo"
        />
        <BulmaFieldInput
          v-model="editing.description"
          label="Description"
          placeholder="Short description"
        />
        <div class="is-flex mb-3" style="gap: 1rem">
          <BulmaField v-slot="{ idForLabel }" label="Width">
            <input
              :id="idForLabel"
              v-model.number="editing.width"
              class="input"
              inputmode="numeric"
            />
          </BulmaField>
          <BulmaField v-slot="{ idForLabel }" label="Height">
            <input
              :id="idForLabel"
              v-model.number="editing.height"
              class="input"
              inputmode="numeric"
            />
          </BulmaField>
        </div>
        <BulmaField
          label="Image URLs"
          help="One image URL per slot; one will be chosen randomly on each page load."
        >
          <SiteParamsFiles
            :files="editing.items"
            :file-props="fileList"
            @add="editing.items.push($event)"
            @remove="editing.items.splice($event, 1)"
          />
        </BulmaField>
        <div class="mt-3 buttons">
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
