<script setup>
import { useSiteParamsEditor } from "@/api/use-site-params-editor.ts";
import { useRoute } from "vue-router";

const route = useRoute();

const editor = useSiteParamsEditor(route.meta.location);
</script>

<template>
  <MetaHead>
    <title>{{ route.meta.title }} • Spotlight PA Almanack</title>
  </MetaHead>
  <ViewSiteParamsBase :title="route.meta.title" :editor="editor">
    <template #form="{ params, fileProps, setRef }">
      <SubsiteParams
        :ref="setRef"
        :params="params"
        :file-props="fileProps"
      ></SubsiteParams>
    </template>
  </ViewSiteParamsBase>
</template>
