<script setup>
import { ref } from "vue";

const props = defineProps({ params: Object, fileProps: Object });

const banner = ref();
const topper = ref();

function saveParams() {
  return {
    schedule_for: props.params.scheduleFor,
    data: {
      ...props.params.data,
      ...banner.value.saveData(),
      ...topper.value.saveData(),
    },
  };
}

defineExpose({
  saveParams,
});
</script>

<template>
  <div>
    <SiteParamsBanner ref="banner" :file-props="fileProps" :params="params" />
    <SiteParamsTopper ref="topper" :file-props="fileProps" :params="params" />
    <!-- <SiteParamsTop :file-props="fileProps" :params="params" />
    <SiteParamsSticky :file-props="fileProps" :params="params" />
    <SiteParamsTakeover :file-props="fileProps" :params="params" />
    <SiteParamsSidebar :file-props="fileProps" :params="params" />
    <SiteParamsSupport :file-props="fileProps" :params="params" />
    <SiteParamsAdFeatured :file-props="fileProps" :params="params" />
    <SiteParamsAdBreaker :file-props="fileProps" :params="params" />
    <SiteParamsAdHeadwater :file-props="fileProps" :params="params" />
    <SiteParamsAdRail :file-props="fileProps" :params="params" />
    <SiteParamsRiver :file-props="fileProps" :params="params" />
    <SiteParamsFeatured :file-props="fileProps" :params="params" />
    <SiteParamsFooter :file-props="fileProps" :params="params" /> -->
  </div>
</template>
