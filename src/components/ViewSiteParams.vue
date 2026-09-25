<script setup>
import { useSiteParamsEditor } from "@/api/use-site-params-editor.ts";

const editor = useSiteParamsEditor("config/_default/params.json");
</script>

<template>
  <MetaHead>
    <title>Sitewide Settings • Spotlight PA Almanack</title>
  </MetaHead>
  <ViewSiteParamsBase
    title="Sitewide Settings"
    :breadcrumb-to="{ name: 'site-params' }"
    :editor="editor"
  >
    <template #form="{ params, fileProps, setRef }">
      <SiteParams
        :ref="setRef"
        :params="params"
        :file-props="fileProps"
      ></SiteParams>
    </template>
  </ViewSiteParamsBase>
</template>
