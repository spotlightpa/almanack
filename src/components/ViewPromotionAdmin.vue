<script setup>
import { ref, reactive, computed } from "vue";

import { get, post, listPromotions, postPromotion } from "@/api/client-v2.js";
import { makeState } from "@/api/service-util.js";
import { useFileList } from "@/api/file-list.js";

// ── file list (for image picker inside editor) ───────────────────────────
const fileList = useFileList();

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
const editing = ref(null);

let itemCounter = 0;

function blankItem() {
  return {
    _id: itemCounter++,
    link: "https://www.spotlightpa.org/donate/",
    label: "",
    labelLink: "",
    description: "",
    sources: [],
  };
}

function deserializeItems(data) {
  // data may be an array (the image-set list) or a legacy object/null
  let arr = Array.isArray(data) ? data : [];
  return arr.map((o) => ({ ...o, _id: itemCounter++ }));
}

function serializeItems(items) {
  // strip the internal _id before sending
  return items.map(({ _id, ...rest }) => rest);
}

function startEdit(promo) {
  editing.value = reactive({
    id: promo.id,
    name: promo.name,
    description: promo.description,
    width: promo.width,
    height: promo.height,
    items: deserializeItems(promo.data),
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

function addItem() {
  editing.value.items.push(reactive(blankItem()));
}

function removeItem(i) {
  editing.value.items.splice(i, 1);
}

async function save() {
  const payload = {
    id: editing.value.id,
    name: editing.value.name,
    description: editing.value.description,
    width: Number(editing.value.width),
    height: Number(editing.value.height),
    data: serializeItems(editing.value.items),
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
        {{ editing.id ? `Edit promotion set #${editing.id}` : "New promotion set" }}
      </h2>

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

      <div class="columns">
        <div class="column is-narrow">
          <BulmaField v-slot="{ idForLabel }" label="Width">
            <input
              :id="idForLabel"
              v-model.number="editing.width"
              class="input"
              inputmode="numeric"
            />
          </BulmaField>
        </div>
        <div class="column is-narrow">
          <BulmaField v-slot="{ idForLabel }" label="Height">
            <input
              :id="idForLabel"
              v-model.number="editing.height"
              class="input"
              inputmode="numeric"
            />
          </BulmaField>
        </div>
      </div>

      <h4 class="title is-5">
        {{ editing.items.length }}
        promotion{{ editing.items.length !== 1 ? "s" : "" }} in set
      </h4>
      <h5 class="subtitle">
        If multiple promotions are in the set, one will be chosen randomly on
        page load.
      </h5>

      <ul>
        <li
          v-for="(item, i) of editing.items"
          :key="item._id"
          class="zebra-row p-3"
        >
          <BulmaFieldInput
            v-model="item.link"
            label="Link URL"
            type="url"
            placeholder="https://www.spotlightpa.org/donate/"
          />
          <BulmaFieldInput
            v-model="item.label"
            label="Banner label"
            placeholder="Sponsored by Acme"
            help="Text accompanying a banner specifying the sponsor or presenter"
          />
          <BulmaFieldInput
            v-model="item.labelLink"
            label="Banner label link"
            type="url"
            placeholder="https://www.spotlightpa.org/support/"
            help="Link that clicking the ad label will go to"
          />
          <BulmaTextarea
            v-model="item.description"
            label="Image description (alt text)"
            help="For blind readers and search engines"
          />
          <BulmaField
            label="Images"
            help="If multiple images are provided, each page load will select one randomly"
          >
            <SiteParamsFiles
              :files="item.sources"
              :file-props="fileList"
              @add="item.sources.push($event)"
              @remove="item.sources.splice($event, 1)"
            />
          </BulmaField>
          <div class="mt-1 mb-2 buttons">
            <button
              type="button"
              class="button is-danger has-text-weight-semibold is-small"
              @click="removeItem(i)"
            >
              Remove promotion from set
            </button>
          </div>
        </li>
      </ul>

      <div class="my-4 buttons">
        <button
          type="button"
          class="button is-success has-text-weight-semibold"
          @click="addItem"
        >
          Add promotion to set
        </button>
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
        <span>New promotion set</span>
      </button>
    </div>

    <!-- ── list ───────────────────────────────────────────────────── -->
    <APILoader
      :is-loading="listState.isLoading"
      :reload="reload"
      :error="listState.error"
    >
      <p v-if="!promotions.length" class="has-text-grey">No promotion sets yet.</p>

      <table v-else class="table is-bordered is-striped is-narrow is-fullwidth">
        <thead>
          <tr>
            <th>ID</th>
            <th>Name</th>
            <th>Description</th>
            <th>Items</th>
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
            <td class="is-narrow">
              {{ Array.isArray(promo.data) ? promo.data.length : 0 }}
            </td>
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
