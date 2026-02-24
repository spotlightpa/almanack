<script setup>
import { watchAPI } from "@/api/service-util.js";
import { get, listPages } from "@/api/client-v2.js";
import PageListItem from "@/api/spotlightpa-page-list-item.js";
import { useRoute } from "vue-router";

const props = defineProps({
  page: { default: "" },
});

const route = useRoute();

const { apiState, fetch, computedList, computedProp } = watchAPI(
  () => [props.page, route.meta.contentPath],
  ([page, path]) => get(listPages, { page, path })
);

const title = route.meta.title;
const pages = computedList("pages", (page) => new PageListItem(page));
const nextPage = computedProp("next_page", (page) => ({
  name: route.name,
  query: { page },
}));
</script>

<template>
  <MetaHead>
    <title>{{ title }} • Spotlight PA Almanack</title>
  </MetaHead>

  <PageList
    :title="title"
    :page="page"
    :next-page="nextPage"
    :api-state="apiState"
    :reload="fetch"
    :pages="pages"
  ></PageList>
</template>
