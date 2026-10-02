<script setup></script>

<template>
  <MetaHead>
    <title>Sitewide Settings • Spotlight PA Almanack</title>
  </MetaHead>
  <SiteParamsBase
    title="Sitewide Settings"
    location="config/_default/params.json"
  >
    <template #form="{ params, fileProps, setRef }">
      <SiteParams
        :ref="setRef"
        :params="params"
        :file-props="fileProps"
      ></SiteParams>
    </template>
  </SiteParamsBase>
</template>
