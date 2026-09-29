<script setup>
import { useRoute } from "vue-router";

const route = useRoute();
</script>

<template>
  <MetaHead>
    <title>{{ route.meta.title }} • Spotlight PA Almanack</title>
  </MetaHead>
  <SiteParamsBase :title="route.meta.title" :location="route.meta.location">
    <template #form="{ params, fileProps, setRef }">
      <SubsiteParams
        :ref="setRef"
        :params="params"
        :file-props="fileProps"
      ></SubsiteParams>
    </template>
  </SiteParamsBase>
</template>
