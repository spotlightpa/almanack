<script setup>
import { watchAPI } from "@/api/service-util.js";
import { get } from "@/api/client-v2.js";

const props = defineProps({
  title: {
    type: String,
    required: true,
  },
  apiPath: {
    type: String,
    required: true,
  },
});

const { apiState, fetch, computedList } = watchAPI(
  () => null,
  () => get(props.apiPath)
);

const pages = computedList("pages", (page) => page);

function swap(event, i) {
  pages.value[i] = event;
}
</script>

<template>
  <MetaHead>
    <title>{{ title }} • Spotlight PA Almanack</title>
  </MetaHead>

  <div>
    <BulmaBreadcrumbs
      :links="[
        { name: 'Admin', to: { name: 'admin' } },
        { name: title, to: {} },
      ]"
    ></BulmaBreadcrumbs>
    <h1 class="title">{{ title }}</h1>

    <APILoader
      :is-loading="apiState.isLoading.value"
      :reload="fetch"
      :error="apiState.error.value"
    >
      <table class="table is-striped is-narrow is-fullwidth">
        <tbody>
          <tr v-for="(page, i) of pages" :key="page.id">
            <td>
              <TaxonomyListRow
                :model-value="page"
                @update:model-value="swap($event, i)"
              />
            </td>
          </tr>
        </tbody>
      </table>
    </APILoader>
  </div>
</template>
