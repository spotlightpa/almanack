<script setup>
import { ref } from "vue";

const props = defineProps({ params: Object, fileProps: Object });

const banner = ref();
const topper = ref();
const top = ref();
const sticky = ref();
const takeover = ref();
const sidebar = ref();
const support = ref();
const adFeatured = ref();
const adBreaker = ref();
const adHeadwater = ref();
const adRail = ref();
const river = ref();
const featured = ref();
const footer = ref();

function saveParams() {
  return {
    schedule_for: props.params.scheduleFor,
    data: {
      ...props.params.data,
      ...banner.value.saveData(),
      ...topper.value.saveData(),
      ...top.value.saveData(),
      ...sticky.value.saveData(),
      ...takeover.value.saveData(),
      ...sidebar.value.saveData(),
      ...support.value.saveData(),
      ...adFeatured.value.saveData(),
      ...adBreaker.value.saveData(),
      ...adHeadwater.value.saveData(),
      ...adRail.value.saveData(),
      ...river.value.saveData(),
      ...featured.value.saveData(),
      ...footer.value.saveData(),
    },
  };
}

defineExpose({
  saveParams,
});
</script>

<template>
  <SiteParamsBanner ref="banner" :file-props="fileProps" :params="params" />
  <SiteParamsTopper ref="topper" :file-props="fileProps" :params="params" />
  <SiteParamsTop ref="top" :file-props="fileProps" :params="params" />
  <SiteParamsSticky ref="sticky" :file-props="fileProps" :params="params" />
  <SiteParamsTakeover ref="takeover" :file-props="fileProps" :params="params" />
  <SiteParamsSidebar ref="sidebar" :file-props="fileProps" :params="params" />
  <SiteParamsSupport ref="support" :file-props="fileProps" :params="params" />
  <SiteParamsAdFeatured
    ref="adFeatured"
    :file-props="fileProps"
    :params="params"
  />
  <SiteParamsAdBreaker
    ref="adBreaker"
    :file-props="fileProps"
    :params="params"
  />
  <SiteParamsAdHeadwater
    ref="adHeadwater"
    :file-props="fileProps"
    :params="params"
  />
  <SiteParamsAdRail ref="adRail" :file-props="fileProps" :params="params" />
  <SiteParamsRiver ref="river" :file-props="fileProps" :params="params" />
  <SiteParamsFeatured ref="featured" :file-props="fileProps" :params="params" />
  <SiteParamsFooter ref="footer" :file-props="fileProps" :params="params" />
</template>
